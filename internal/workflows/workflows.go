package workflows

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

func GenerateReport(repos []string) Report {
	return Report{}
}

func getClient(jwt string) (context.Context, *github.Client) {
	ctx := context.Background()

	// Create OAuth2 token source
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: jwt},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Create GitHub client
	return ctx, github.NewClient(tc)
}

// FindWorkflows loads all the GitHub Actions workflows for a given repository.
// jwt is the JSON Web Token used for authentication.
// repo is the repository in the format "owner/repo".
func FindWorkflows(jwt string, repo string) []string {
	// Create GitHub client
	ctx, client := getClient(jwt)

	// Split repo into owner and name
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return []string{}
	}
	owner, repoName := parts[0], parts[1]

	// List contents of .github/workflows directory
	_, dirContent, _, err := client.Repositories.GetContents(ctx, owner, repoName, ".github/workflows", nil)
	if err != nil {
		return []string{}
	}

	var workflows []string

	// Iterate through files and fetch YAML/YML files
	for _, content := range dirContent {
		if content.GetType() == "file" {
			name := content.GetName()
			if strings.HasSuffix(strings.ToLower(name), ".yml") || strings.HasSuffix(strings.ToLower(name), ".yaml") {
				workflows = append(workflows, name)
			}
		}
	}

	return workflows
}

func WorkflowToString(jwt string, repo string, workflow string) string {
	// Create GitHub client
	ctx, client := getClient(jwt)

	// Split repo into owner and name
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return ""
	}
	owner, repoName := parts[0], parts[1]

	// Get file content
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repoName, workflow, nil)
	if err != nil {
		return ""
	}

	// Decode content
	contentStr, err := fileContent.GetContent()
	if err != nil {
		return ""
	}

	return contentStr
}

func FindContributorsToWorkflow(jwt string, repo string, workflow string) []string {
	// Create GitHub client
	ctx, client := getClient(jwt)

	// Split repo into owner and name
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return []string{}
	}
	owner, repoName := parts[0], parts[1]

	// Construct the workflow file path
	workflowPath := ".github/workflows/" + workflow

	// Get commits for the specific workflow file
	opts := &github.CommitsListOptions{
		Path: workflowPath,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Track unique contributors
	contributorsMap := make(map[string]bool)
	var contributors []string

	// Fetch all commits for the workflow file (handle pagination)
	for {
		commits, resp, err := client.Repositories.ListCommits(ctx, owner, repoName, opts)
		if err != nil {
			return []string{}
		}

		// Extract unique contributor names
		for _, commit := range commits {
			if commit.Commit != nil && commit.Commit.Author != nil && commit.Commit.Author.Name != nil {
				name := *commit.Commit.Author.Name
				if !contributorsMap[name] {
					contributorsMap[name] = true
					contributors = append(contributors, name)
				}
			}
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return contributors
}

// ParseWorkflow parses the string representation of a GitHub Actions workflow
// and returns a slice of Action structs representing the actions used in the workflow.
func ParseWorkflow(workflow string) []Action {
	var workflowMap map[string]interface{}

	err := yaml.Unmarshal([]byte(workflow), &workflowMap)
	if err != nil {
		return []Action{}
	}

	// Extract jobs
	jobsInterface, ok := workflowMap["jobs"]
	if !ok {
		return []Action{}
	}

	jobsMap, ok := jobsInterface.(map[string]interface{})
	if !ok {
		return []Action{}
	}

	var actions []Action

	// Iterate through jobs and parse actions
	for _, jobInterface := range jobsMap {
		jobMap, ok := jobInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract steps
		stepsInterface, ok := jobMap["steps"]
		if !ok {
			continue
		}

		stepsSlice, ok := stepsInterface.([]interface{})
		if !ok {
			continue
		}

		for _, stepInterface := range stepsSlice {
			stepMap, ok := stepInterface.(map[string]interface{})
			if !ok {
				continue
			}

			usesInterface, ok := stepMap["uses"]
			if !ok {
				continue
			}

			uses, ok := usesInterface.(string)
			if !ok {
				continue
			}

			// Split uses into action and version
			var actionName, actionVersion string
			if strings.Contains(uses, "@") {
				parts := strings.SplitN(uses, "@", 2)
				actionName = parts[0]
				actionVersion = parts[1]
			} else {
				actionName = uses
				actionVersion = "latest"
			}

			env := convertStringMap(getChildMap(stepMap, "env"))
			with := convertStringMap(getChildMap(stepMap, "with"))

			action := Action{
				Uses:        actionName,
				UsesVersion: actionVersion,
				Settings:    make(map[string]string),
				Env:         env,
				With:        with,
			}

			actions = append(actions, action)
		}
	}

	return actions
}

func getChildMap(input map[string]interface{}, key string) map[string]interface{} {
	childInterface, ok := input[key]
	if !ok {
		return nil
	}
	childMap, ok := childInterface.(map[string]interface{})
	if !ok {
		return nil
	}
	return childMap
}

func convertStringMap(input map[string]interface{}) map[string]string {
	if input == nil {
		return nil
	}

	result := make(map[string]string)
	for key, value := range input {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}

func FindActionsWithDifferentVersions(actions1 []Action, actions2 []Action) int {
	return 0
}

func FindActionsWithSimilarConfigurations(actions1 []Action, actions2 []Action) int {
	return 0
}
