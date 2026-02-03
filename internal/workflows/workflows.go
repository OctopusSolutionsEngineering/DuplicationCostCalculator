package workflows

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

const HighSimilarity = 30

func GenerateReport(jwt string, repos []string) Report {
	workflowsContent := make(map[string][]string)

	for _, repo := range repos {
		workflowFiles := FindWorkflows(jwt, repo)

		for _, workflowFile := range workflowFiles {
			workflowStr := WorkflowToString(jwt, repo, workflowFile)
			if workflowStr != "" {
				workflowsContent[repo] = append(workflowsContent[repo], workflowStr)
			}
		}
	}

	return GenerateReportFromWorkflows(workflowsContent)
}

func GenerateReportFromWorkflows(workflows map[string][]string) Report {

	repoActions := convertWorkflowToActionsMap(workflows)

	repoNames := maps.Keys(repoActions)
	sortedRepoNames := slices.Sorted(repoNames)

	report := Report{
		Comparisons: map[string]map[string]RepoMeasurements{},
	}

	for i := 0; i < len(sortedRepoNames); i++ {
		for j := i + 1; j < len(sortedRepoNames); j++ {
			repo1 := sortedRepoNames[i]
			repo2 := sortedRepoNames[j]

			actionsList1 := repoActions[repo1]
			actionsList2 := repoActions[repo2]

			for _, actions1 := range actionsList1 {
				for _, actions2 := range actionsList2 {
					diffVersions := FindActionsWithDifferentVersions(actions1, actions2)
					similarConfigs := FindActionsWithSimilarConfigurations(actions1, actions2)

					if _, ok := report.Comparisons[repo1]; !ok {
						report.Comparisons[repo1] = make(map[string]RepoMeasurements)
					}

					if _, ok := report.Comparisons[repo2]; !ok {
						report.Comparisons[repo2] = make(map[string]RepoMeasurements)
					}

					report.Comparisons[repo1][repo2] = RepoMeasurements{
						StepsWithDifferentVersions: diffVersions,
						StepsWithSimilarConfig:     similarConfigs,
					}

					report.Comparisons[repo2][repo1] = RepoMeasurements{
						StepsWithDifferentVersions: diffVersions,
						StepsWithSimilarConfig:     similarConfigs,
					}

					if diffVersions > 0 || similarConfigs > 0 {
						fmt.Printf("Comparison between %s and %s:\n", repo1, repo2)
						if diffVersions > 0 {
							fmt.Printf("  Steps with different versions: %d\n", diffVersions)
						}
						if similarConfigs > 0 {
							fmt.Printf("  Steps with similar configurations: %d\n", similarConfigs)
						}
					}
				}
			}
		}
	}

	return report
}

func convertWorkflowToActionsMap(workflows map[string][]string) map[string][][]Action {
	repoActions := make(map[string][][]Action)

	for repo, workflowFiles := range workflows {
		fmt.Printf("Repo: %s\n", repo)
		for _, workflowFile := range workflowFiles {
			fmt.Printf("  Workflow: %s\n", workflowFile)

			actions := ParseWorkflow(workflowFile)

			for _, action := range actions {
				fmt.Printf("    Action: %s@%s\n", action.Uses, action.UsesVersion)
			}

			repoActions[repo] = append(repoActions[repo], actions)
		}
	}

	return repoActions
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

	// 1. Get all keys into a slice
	keys := slices.Collect(maps.Keys(jobsMap))

	// 2. Sort the keys alphabetically
	slices.Sort(keys)

	var actions []Action

	// Iterate through jobs and parse actions
	for _, key := range keys {
		jobMap, ok := jobsMap[key].(map[string]interface{})
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

			uses := ""
			usesInterface, ok := stepMap["uses"]
			if ok {
				usesString, ok := usesInterface.(string)
				if ok {
					// Script steps often don't have a 'uses' field
					uses = usesString
				}
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
			settings := getOtherValues(stepMap, []string{"uses", "env", "with"})

			action := Action{
				Uses:        actionName,
				UsesVersion: actionVersion,
				Settings:    settings,
				Env:         env,
				With:        with,
			}

			action.GenerateHash()

			actions = append(actions, action)
		}
	}

	return actions
}

func getOtherValues(input map[string]interface{}, excludeKeys []string) map[string]string {
	result := make(map[string]string)
	excludeMap := make(map[string]bool)
	for _, key := range excludeKeys {
		excludeMap[key] = true
	}

	for key, value := range input {
		if !excludeMap[key] {
			result[key] = fmt.Sprintf("%v", value)
		}
	}

	return result
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
	count := 0

	for _, action1 := range actions1 {
		for _, action2 := range actions2 {
			if action1.Uses == action2.Uses && action1.UsesVersion != action2.UsesVersion {
				count += 2
			}
		}
	}

	return count
}

func FindActionsWithSimilarConfigurations(actions1 []Action, actions2 []Action) int {

	count := 0

	for _, action1 := range actions1 {
		for _, action2 := range actions2 {
			if action1.Uses == action2.Uses {
				if action1.hash != nil && action2.hash != nil {
					distance := action1.hash.Diff(action2.hash)

					if distance <= HighSimilarity {
						// two steps that are similar
						count += 2
					}
				}
			}
		}
	}

	return count
}
