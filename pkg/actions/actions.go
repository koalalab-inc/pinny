package actions

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
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
		// Check for exact tag match
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

	// Check for shortened hash
	if exactRef == nil {
		for _, r := range refs {
			sha := *r.GetObject().SHA
			refType := r.GetObject().GetType()
			if refType == "commit" && strings.HasPrefix(sha, ref) {
				if sha != ref {
					fmt.Printf("WARN:: Shortened hash found for ref %s/%s@%s.s\nIt is recommended to use full 40 character hash.", owner, repo, ref)
				}
				exactRef = r
				break
			}
		}
	}

	//check for impostor commits
	if exactRef == nil {
		fmt.Printf("WARN:: No exact match found for ref %s/%s@%s\n", owner, repo, ref)
		impostor := true
		if exactRefType != "tag" && exactRefType != "branch" {
			for _, r := range refs {
				rRef := r.GetRef()
				if strings.HasPrefix(rRef, "refs/tags/") || strings.HasPrefix(rRef, "refs/heads/") {
					contained, err := refContains(ctx, client, owner, repo, rRef, ref)
					if err != nil {
						return nil, nil, err
					}
					if contained {
						impostor = false
						break
					}
				}
			}
			if impostor {
				fmt.Printf("WARN:: Impostor found for ref %s/%s@%s\n", owner, repo, ref)
			}
		}
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
		for _, r := range refs {
			if *r.GetObject().SHA == *exactRef.GetObject().SHA && r.GetRef() != exactRef.GetRef() {
				otherMatchingRefs = append(otherMatchingRefs, r)
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
	githubActionRef, err := parseActionString(actionString)
	if err != nil {
		return nil, err
	}
	owner := githubActionRef.Owner
	repo := githubActionRef.Repo
	ref := githubActionRef.Ref

	cacheKey := fmt.Sprintf("%s/%s@%s", owner, repo, ref)
	if githubActionRef, ok := actionRefCache[cacheKey]; ok {
		return githubActionRef, nil
	}

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

	actionRefCache[cacheKey] = githubActionRef
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

func refContains(ctx context.Context, c *github.Client, owner, repo, base, target string) (bool, error) {
	diff, resp, err := c.Repositories.CompareCommits(ctx, owner, repo, base, target, &github.ListOptions{PerPage: 1})
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			// NotFound can be returned for some divergent cases: "404 No common ancestor between ..."
			return false, nil
		}
		return false, fmt.Errorf("error comparing revisions: %w", err)
	}

	// Target should be behind or at the base ref if it is considered contained.
	return diff.GetStatus() == "behind" || diff.GetStatus() == "identical", nil
}
