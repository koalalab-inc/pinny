package actions

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/koalalab-inc/pinny/pkg/docker"
	"github.com/koalalab-inc/pinny/pkg/utils"

	"github.com/google/go-github/v56/github"
)

const workflowDir = ".github/workflows"

var actionRefCache = make(map[string]*GithubActionRef)

func getTokenFromEnv() *string {
	token, exists := os.LookupEnv("GITHUB_TOKEN")
	token = strings.TrimSpace(token)
	if !exists || token == "" {
		return nil
	}
	return &token
}

func getGithubClient(token *string) *github.Client {
	client := github.NewClient(nil)
	if token != nil {
		return client.WithAuthToken(*token)
	}
	return client
}

type GithubActionRef struct {
	Raw           string
	Digest        string
	Owner         string
	Repo          string
	Path          string
	Ref           string
	OtherRefNames []string
}

func (g *GithubActionRef) NameWithRef() string {
	name := fmt.Sprintf("%s/%s", g.Owner, g.Repo)
	if g.Path != "" {
		name = fmt.Sprintf("%s/%s", name, g.Path)
	}
	name = fmt.Sprintf("%s@%s", name, g.Ref)
	return name
}

func (g *GithubActionRef) NameWithDigest() string {
	name := fmt.Sprintf("%s/%s", g.Owner, g.Repo)
	if g.Path != "" {
		name = fmt.Sprintf("%s/%s", name, g.Path)
	}
	name = fmt.Sprintf("%s@%s", name, g.Digest)
	return name
}

func parseActionString(actionString string) (*GithubActionRef, error) {
	githubRepoActionRegex := regexp.MustCompile(`(?P<owner>[^/]+)/(?P<repo>[^@/]+)(/(?P<path>[^@]+))?@(?P<ref>.+)`)
	if ok, matches := utils.MatchNamedRegex(githubRepoActionRegex, actionString); ok {
		return &GithubActionRef{
			Raw:   actionString,
			Owner: matches["owner"],
			Repo:  matches["repo"],
			Path:  matches["path"],
			Ref:   matches["ref"],
		}, nil
	} else {
		return nil, fmt.Errorf("invalid action string")
	}
}

func getActionDigest(owner string, repo string, ref string) (*string, []*github.Reference, error) {
	token := getTokenFromEnv()
	client := getGithubClient(token)
	ctx := context.Background()
	tagRef := fmt.Sprintf("tags/%s", ref)
	branchRef := fmt.Sprintf("heads/%s", ref)
	var digest string
	opts := &github.ReferenceListOptions{
		Ref: "",
	}

	refs, _, err := client.Git.ListMatchingRefs(ctx, owner, repo, opts)
	if err != nil {
		return nil, nil, err
	}

	var exactRef *github.Reference
	var exactRefType string
	for _, ref := range refs {
		if ref.GetRef() == fmt.Sprintf("refs/%s", tagRef) {
			exactRef = ref
			exactRefType = "tag"
			break
		} else if ref.GetRef() == fmt.Sprintf("refs/%s", branchRef) {
			exactRef = ref
			exactRefType = "branch"
			break
		}
	}

	if exactRefType == "branch" {
		fmt.Printf("WARN:: Branch references are being used for third party Github Action: %s/%s@%s\n", owner, repo, ref)
	}

	if exactRef == nil {
		opts = &github.ReferenceListOptions{
			Ref: "",
		}
		refs, _, err := client.Git.ListMatchingRefs(ctx, owner, repo, opts)
		if err != nil {
			return nil, nil, err
		}
		for _, r := range refs {
			sha := *r.GetObject().SHA
			refType := r.GetObject().GetType()
			if refType == "commit" && strings.Contains(sha, ref) {
				exactRef = r
				break
			}
		}
	}

	if exactRef == nil {
		fmt.Printf("WARN:: No exact match found for ref %s/%s@%s\n", owner, repo, ref)
		return &ref, []*github.Reference{}, nil
	}
	refObjectType := exactRef.GetObject().GetType()
	if refObjectType == "tag" {
		tag, _, err := client.Git.GetTag(ctx, owner, repo, *exactRef.GetObject().SHA)
		if err != nil {
			return nil, nil, err
		}
		digest = *tag.GetObject().SHA
	} else {
		digest = *exactRef.GetObject().SHA
	}

	otherMatchingRefs := []*github.Reference{}
	if exactRef != nil {
		for _, ref := range refs {
			if *ref.GetObject().SHA == *exactRef.GetObject().SHA && ref.GetRef() != exactRef.GetRef() {
				otherMatchingRefs = append(otherMatchingRefs, ref)
			}
		}
	}

	return &digest, otherMatchingRefs, nil
}

func GetDigest(actionString string) (*string, error) {
	githubActionRef, err := parseActionString(actionString)
	if err != nil {
		return nil, err
	}
	owner := githubActionRef.Owner
	repo := githubActionRef.Repo
	ref := githubActionRef.Ref

	digest, _, err := getActionDigest(owner, repo, ref)
	if err != nil {
		return nil, err
	}

	githubActionRef.Digest = *digest

	return digest, nil
}

func GetGithubActionRefWithDigest(actionString string) (*GithubActionRef, error) {
	if githubActionRef, ok := actionRefCache[actionString]; ok {
		return githubActionRef, nil
	}

	githubActionRef, err := parseActionString(actionString)
	if err != nil {
		return nil, err
	}
	owner := githubActionRef.Owner
	repo := githubActionRef.Repo
	ref := githubActionRef.Ref

	digest, matchingRefs, err := getActionDigest(owner, repo, ref)
	otherRefNamesArr := []string{}
	for _, ref := range matchingRefs {
		refName := ref.GetRef()
		if refNameArr := strings.Split(ref.GetRef(), "/"); len(refNameArr) > 2 {
			refName = refNameArr[2]
		}
		otherRefNamesArr = append(otherRefNamesArr, refName)
	}

	if err != nil {
		return nil, err
	}

	githubActionRef.Digest = *digest
	githubActionRef.OtherRefNames = otherRefNamesArr

	actionRefCache[actionString] = githubActionRef
	return githubActionRef, nil
}

func PinWorkflow(workflowName string) error {
	workflow, err := os.Open(fmt.Sprintf("%s/%s", workflowDir, workflowName))
	if err != nil {
		return err
	}

	tmpWorkflow, err := os.OpenFile(fmt.Sprintf("%s/%s.tmp", workflowDir, workflowName), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	tmpWorkflowWriter := bufio.NewWriter(tmpWorkflow)

	defer tmpWorkflow.Close()
	defer tmpWorkflowWriter.Flush()

	workflowScanner := bufio.NewScanner(workflow)

	usesDockerRegex := regexp.MustCompile(`^(?P<pre>.*uses\s*:\s*)(?P<actionString>docker://\S+)(?P<post>.*)$`)
	usesActionRegex := regexp.MustCompile(`^(?P<pre>.*uses\s*:\s*)(?P<actionString>\S+@\S+)(?P<post>.*)$`)

	for workflowScanner.Scan() {
		line := workflowScanner.Text()
		if strings.Contains(line, "uses:") {
			var actionString string
			if ok, matches := utils.MatchNamedRegex(usesDockerRegex, line); ok {
				actionString = matches["actionString"]
				dockerImageRef, err := docker.GetImageRefWithDigest(actionString)
				if err != nil {
					return err
				}
				pinnedActionString := fmt.Sprintf("docker://%s", dockerImageRef.OriginalName("digest"))
				if pinnedActionString == actionString {
					tmpWorkflowWriter.WriteString(fmt.Sprintf("%s\n", line))
					continue
				}
				comment := fmt.Sprintf(" # %s", dockerImageRef.Raw)
				tmpWorkflowWriter.WriteString(fmt.Sprintf("%s%s%s\n", matches["pre"], pinnedActionString, comment))
			} else if ok, matches := utils.MatchNamedRegex(usesActionRegex, line); ok {
				actionString = matches["actionString"]
				githubActionRef, err := GetGithubActionRefWithDigest(actionString)
				if err != nil {
					return err
				}
				pinnedActionString := githubActionRef.NameWithDigest()
				if pinnedActionString == actionString {
					tmpWorkflowWriter.WriteString(fmt.Sprintf("%s\n", line))
					continue
				}
				comment := fmt.Sprintf(" # %s", githubActionRef.Raw)
				if otherNamesArr := githubActionRef.OtherRefNames; len(otherNamesArr) > 0 {
					otherNames := strings.Join(otherNamesArr, ",")
					comment = fmt.Sprintf("%s | %s", comment, otherNames)
				}
				tmpWorkflowWriter.WriteString(fmt.Sprintf("%s%s%s\n", matches["pre"], pinnedActionString, comment))
			} else {
				tmpWorkflowWriter.WriteString(fmt.Sprintf("%s\n", line))
			}
		} else {
			tmpWorkflowWriter.WriteString(fmt.Sprintf("%s\n", line))
		}
	}
	return nil
}
