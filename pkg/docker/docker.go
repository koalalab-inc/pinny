package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/koalalab-inc/pinny/pkg/utils"

	"github.com/asottile/dockerfile"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/transports/alltransports"
)

const Lockfile = "pinny-lock.json"

type FromCmd struct {
	Flags []string        `json:"flags"`
	Image *DockerImageRef `json:"image"`
	Alias string          `json:"alias"`
}

type DockerImageRef struct {
	Raw    string `json:"raw"`
	Host   string `json:"host"`
	Repo   string `json:"repo"`
	Name   string `json:"name"`
	Tag    string `json:"tag"`
	Digest string `json:"digest"`
}

func (d *DockerImageRef) withSuffix(suffix string) string {
	if suffix == "tag" && d.Tag != "" {
		return fmt.Sprintf(":%s", d.Tag)
	} else if suffix == "digest" && d.Digest != "" {
		return fmt.Sprintf("@%s", d.Digest)
	} else {
		return ""
	}
}

func (d *DockerImageRef) fullName(suffix string) string {
	host := d.Host
	if host == "" {
		host = "docker.io"
	}
	repo := d.Repo
	if repo == "" && host == "docker.io" {
		repo = "library"
	}
	name := d.Name
	imageName := fmt.Sprintf("docker://%s/%s/%s", host, repo, name)
	return fmt.Sprintf("%s%s", imageName, d.withSuffix(suffix))
}

func (d *DockerImageRef) OriginalName(suffix string) string {
	imageName := d.Name
	repo := d.Repo
	if repo != "" {
		imageName = fmt.Sprintf("%s/%s", repo, imageName)
	}
	host := d.Host
	if host != "" {
		imageName = fmt.Sprintf("%s/%s", host, imageName)
	}
	return fmt.Sprintf("%s%s", imageName, d.withSuffix(suffix))
}

func (f *FromCmd) stringify(suffix string) string {
	platformString := ""
	if len(f.Flags) > 0 {
		platformString = fmt.Sprintf("%s ", strings.Join(f.Flags[:], " "))
	}
	imageString := f.Image.OriginalName(suffix)
	aliasString := ""
	if f.Alias != "" {
		aliasString = fmt.Sprintf("AS %s", f.Alias)
	}
	resp := fmt.Sprintf("FROM %s%s %s", platformString, imageString, aliasString)
	return resp
}

func getDigest(imageName string) (*string, error) {
	ref, err := alltransports.ParseImageName(imageName)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	digest, err := docker.GetDigest(ctx, nil, ref)
	if err != nil {
		return nil, err
	}

	digestStr := string(digest)
	return &digestStr, nil
}

func GetDigest(imageString string) (*string, error) {
	imageRef, err := getImageRefFromImageString(imageString)
	if err != nil {
		return nil, err
	}

	var imageName string
	if imageRef.Digest != "" {
		imageName = imageRef.fullName("digest")
	} else {
		imageName = imageRef.fullName("tag")
	}

	return getDigest(imageName)
}

func GetImageRefWithDigest(imageString string) (*DockerImageRef, error) {
	imageRef, err := getImageRefFromImageString(imageString)
	if err != nil {
		return nil, err
	}

	var imageName string
	if imageRef.Digest != "" {
		imageName = imageRef.fullName("digest")
	} else {
		imageName = imageRef.fullName("tag")
	}

	digest, err := getDigest(imageName)
	if err != nil {
		return nil, err
	}

	imageRef.Digest = *digest

	return imageRef, nil
}

func GetManifest(imageName string) ([]byte, error) {
	ref, err := alltransports.ParseImageName(imageName)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	img, err := ref.NewImageSource(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer img.Close()
	b, _, err := img.GetManifest(ctx, nil)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GeneratePinnedDockerfile(filename string, offline bool) error {
	var imageDigestMap = make(map[string]string)
	if offline {
		lockFileContents, err := os.ReadFile(Lockfile)
		if err != nil {
			return err
		}
		json.Unmarshal(lockFileContents, &imageDigestMap)
	}

	timestampStr := time.Now().Format(time.RFC1123)

	srcFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	destFilename := fmt.Sprintf("%s.pinned.tmp", filename)
	destFile, err := os.OpenFile(destFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	destFileWriter := bufio.NewWriter(destFile)

	defer srcFile.Close()
	defer destFile.Close()
	defer destFileWriter.Flush()

	srcLines := []string{}

	srcFileScanner := bufio.NewScanner(srcFile)

	for srcFileScanner.Scan() {
		srcLines = append(srcLines, srcFileScanner.Text())
	}

	commands, err := dockerfile.ParseFile(filename)
	if err != nil {
		return err
	}

	startLine := 1

	for _, cmd := range commands {
		if cmd.Cmd == "FROM" {
			// copy lines preceding FROM command to destination from source file
			if cmd.StartLine > startLine {
				for i := startLine; i < cmd.StartLine-1; i++ {
					destFileWriter.WriteString(srcLines[i-1] + "\n")
				}
				if cmd.StartLine > 1 {
					lineBeforeFromCmd := srcLines[cmd.StartLine-2]
					if !strings.HasPrefix(lineBeforeFromCmd, "# Pinned") {
						destFileWriter.WriteString(lineBeforeFromCmd + "\n")
					}
				}
			}
			startLine = cmd.EndLine + 1
			imageString, aliasString := getImageAndAliasFromCmd(cmd)

			imageRef, err := getImageRefFromImageString(imageString)
			if err != nil {
				return err
			}

			if imageRef.Digest != "" {
				if cmd.StartLine > 1 {
					lineBeforeFromCmd := srcLines[cmd.StartLine-2]
					if strings.HasPrefix(lineBeforeFromCmd, "# Pinned") {
						destFileWriter.WriteString(lineBeforeFromCmd + "\n")
					}
				}
				for i := cmd.StartLine; i <= cmd.EndLine; i++ {
					destFileWriter.WriteString(srcLines[i-1] + "\n")
				}
			} else {
				var digest *string
				if offline {
					if d, ok := imageDigestMap[imageRef.fullName("tag")]; ok {
						digest = &d
					} else {
						return fmt.Errorf("digest not found for %s", imageRef.fullName("tag"))
					}
				} else {
					digest, err = getDigest(imageRef.fullName("tag"))
					if err != nil {
						return err
					}
				}
				imageRef.Digest = *digest

				fromCmd := &FromCmd{
					Flags: cmd.Flags,
					Image: imageRef,
					Alias: aliasString,
				}

				commentString := fmt.Sprintf("# Pinned %s using pinny", imageRef.Raw)

				if imageRef.Tag == "" || imageRef.Tag == "latest" {
					commentString = fmt.Sprintf("%s on %s\n", commentString, timestampStr)
				} else {
					commentString = fmt.Sprintf("%s\n", commentString)
				}

				destFileWriter.WriteString(commentString)
				destFileWriter.WriteString(fromCmd.stringify("digest") + "\n")
			}
		}
	}

	// copy remaining lines to destination from source file
	if startLine < len(srcLines) {
		for i := startLine; i <= len(srcLines); i++ {
			destFileWriter.WriteString(srcLines[i-1] + "\n")
		}
	}

	// manifestRespFileMapJSON, _ := json.MarshalIndent(manifestRespFileMap, "", "    ")

	// lockFileWriter.Write(manifestRespFileMapJSON)

	return nil
}

func GeneratePinnyLockFile(filename string) error {
	var imageDigestMap = make(map[string]string)
	lockFileContents, err := os.ReadFile(Lockfile)
	if !os.IsNotExist(err) {
		if err != nil {
			return err
		}
		json.Unmarshal(lockFileContents, &imageDigestMap)
	}

	tmpLockFile := fmt.Sprintf("%s.tmp", Lockfile)
	lockFile, err := os.OpenFile(tmpLockFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	lockFileWriter := bufio.NewWriter(lockFile)

	defer lockFile.Close()
	defer lockFileWriter.Flush()

	commands, err := dockerfile.ParseFile(filename)
	if err != nil {
		return err
	}

	for _, cmd := range commands {
		if cmd.Cmd == "FROM" {
			imageString, _ := getImageAndAliasFromCmd(cmd)

			imageRef, err := getImageRefFromImageString(imageString)
			if err != nil {
				return err
			}

			imageRefString := imageRef.fullName("tag")

			digest, err := getDigest(imageRefString)
			if err != nil {
				return err
			}

			imageDigestMap[imageRefString] = *digest
		}
	}

	imageDigestMap["generated_at"] = time.Now().Format(time.RFC1123)
	imageDigestMap["generated_by"] = "Pinny"

	imageDigestMapJSON, err := json.MarshalIndent(imageDigestMap, "", "    ")
	if err != nil {
		return err
	}

	_, err = lockFileWriter.Write(imageDigestMapJSON)
	if err != nil {
		return err
	}
	err = os.Rename(tmpLockFile, Lockfile)
	return err
}

func getImageRefFromImageString(imageString string) (*DockerImageRef, error) {
	imageWithHost := "((?P<host>[^/]+)/(?P<owner>[^/]+)/(?P<image>[^:]+))"
	imageWithoutHost := "(((?P<owner>[^/]+)/)?(?P<image>[^:@]+))"
	dockerRegexString := fmt.Sprintf("^(docker://)?(%s|%s)(:(?P<tag>.+))?(@(?P<digest>sha.+))?$", imageWithHost, imageWithoutHost)
	dockerRegex := regexp.MustCompile(dockerRegexString)
	if ok, matches := utils.MatchNamedRegex(dockerRegex, imageString); ok {
		imageRef := &DockerImageRef{
			Raw:    imageString,
			Host:   matches["host"],
			Repo:   matches["owner"],
			Name:   matches["image"],
			Tag:    matches["tag"],
			Digest: matches["digest"],
		}
		return imageRef, nil
	} else {
		return nil, fmt.Errorf("invalid image string")
	}
}

func getImageAndAliasFromCmd(cmd dockerfile.Command) (string, string) {
	imageString := ""
	aliasString := ""
	if len(cmd.Value) == 1 {
		imageString = strings.TrimSpace(cmd.Value[0])
		aliasString = ""
	} else if len(cmd.Value) == 3 {
		imageString = strings.TrimSpace(cmd.Value[0])
		aliasString = strings.TrimSpace(cmd.Value[2])
	}
	return imageString, aliasString
}
