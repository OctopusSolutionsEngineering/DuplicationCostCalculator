package workflows

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v57/github"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

const HighSimilarity = 30

func GenerateReport(client *github.Client, repos []string) Report {
	workflowsContent := make(map[string][]string)

	for _, repo := range repos {
		workflowFiles := FindWorkflows(client, repo)

		for _, workflowFile := range workflowFiles {
			workflowStr := WorkflowToString(client, repo, workflowFile)
			if workflowStr != "" {
				workflowsContent[repo] = append(workflowsContent[repo], workflowStr)
			}
		}
	}

	report := GenerateReportFromWorkflows(workflowsContent)

	return report
}

func GenerateReportFromWorkflows(workflows map[string][]string) Report {

	repoActions := convertWorkflowToActionsMap(workflows)

	repoNames := maps.Keys(repoActions)
	sortedRepoNames := slices.Sorted(repoNames)

	report := Report{
		Comparisons:   map[string]map[string]RepoMeasurements{},
		NumberOfRepos: len(sortedRepoNames),
	}

	for i := 0; i < len(sortedRepoNames); i++ {
		for j := i + 1; j < len(sortedRepoNames); j++ {
			repo1 := sortedRepoNames[i]
			repo2 := sortedRepoNames[j]

			actionsList1 := repoActions[repo1]
			actionsList2 := repoActions[repo2]

			for _, actions1 := range actionsList1 {
				for _, actions2 := range actionsList2 {
					diffVersionsActions, diffVersionsCount, diffVersions := FindActionsWithDifferentVersions(actions1, actions2)
					similarConfigsActions, similarConfigsCount, similarConfigs := FindActionsWithSimilarConfigurations(actions1, actions2)

					allActionsIds := lo.Map(append(diffVersionsActions, similarConfigsActions...), func(item Action, index int) string {
						return item.Id
					})

					uniqueActions := lo.Uniq(allActionsIds)

					if _, ok := report.Comparisons[repo1]; !ok {
						report.Comparisons[repo1] = make(map[string]RepoMeasurements)
					}

					if _, ok := report.Comparisons[repo2]; !ok {
						report.Comparisons[repo2] = make(map[string]RepoMeasurements)
					}

					report.Comparisons[repo1][repo2] = RepoMeasurements{
						StepsWithDifferentVersions:       lo.Uniq(append(report.Comparisons[repo1][repo2].StepsWithDifferentVersions, diffVersions...)),
						StepsWithDifferentVersionsCount:  report.Comparisons[repo1][repo2].StepsWithDifferentVersionsCount + diffVersionsCount,
						StepsWithSimilarConfig:           lo.Uniq(append(report.Comparisons[repo1][repo2].StepsWithSimilarConfig, similarConfigs...)),
						StepsWithSimilarConfigCount:      report.Comparisons[repo1][repo2].StepsWithSimilarConfigCount + similarConfigsCount,
						StepsThatIndicateDuplicationRisk: report.Comparisons[repo1][repo2].StepsThatIndicateDuplicationRisk + len(uniqueActions),
					}

					report.Comparisons[repo2][repo1] = RepoMeasurements{
						StepsWithDifferentVersions:       lo.Uniq(append(report.Comparisons[repo2][repo1].StepsWithDifferentVersions, diffVersions...)),
						StepsWithDifferentVersionsCount:  report.Comparisons[repo2][repo1].StepsWithDifferentVersionsCount + diffVersionsCount,
						StepsWithSimilarConfig:           lo.Uniq(append(report.Comparisons[repo2][repo1].StepsWithSimilarConfig, similarConfigs...)),
						StepsWithSimilarConfigCount:      report.Comparisons[repo2][repo1].StepsWithSimilarConfigCount + similarConfigsCount,
						StepsThatIndicateDuplicationRisk: report.Comparisons[repo2][repo1].StepsThatIndicateDuplicationRisk + len(uniqueActions),
					}
				}
			}
		}
	}

	numberOfReposWithDuplicationOrDrift := 0
	for _, repo := range sortedRepoNames {
		foundDuplicationOrDrift := false
		for _, measurements := range report.Comparisons[repo] {
			if measurements.StepsThatIndicateDuplicationRisk > 0 {
				foundDuplicationOrDrift = true

			}
		}
		if foundDuplicationOrDrift {
			numberOfReposWithDuplicationOrDrift++
		}
	}

	report.NumberOfReposWithDuplicationOrDrift = numberOfReposWithDuplicationOrDrift

	return report
}

func convertWorkflowToActionsMap(workflows map[string][]string) map[string][][]Action {
	repoActions := make(map[string][][]Action)

	workflowId := 0
	for repo, workflowFiles := range workflows {
		for _, workflowFile := range workflowFiles {
			workflowId++
			actions := ParseWorkflow(workflowFile, workflowId)
			repoActions[repo] = append(repoActions[repo], actions)
		}
	}

	return repoActions
}

func GetClient(jwt string) *github.Client {
	ctx := context.Background()

	// Create a token source with the JWT
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: jwt},
	)

	// Create an HTTP client with the token
	tc := oauth2.NewClient(ctx, ts)

	// Create and return the GitHub client
	return github.NewClient(tc)
}

func GetClientLocal() *github.Client {

	// Load environment variables for security
	appIDStr := os.Getenv("GITHUB_APP_ID")
	installationIDStr := os.Getenv("GITHUB_INSTALLATION_ID")
	privateKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")

	appID := mustParseInt64(appIDStr)
	installationID := mustParseInt64(installationIDStr)

	// Create an http.RoundTripper that signs requests as a GitHub App installation
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyPath)
	if err != nil {
		log.Fatalf("Error creating ghinstallation transport: %v", err)
	}

	// Create the GitHub client with the authenticated transport
	client := github.NewClient(&http.Client{Transport: itr})

	// Create GitHub client
	return client
}

func mustParseInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Invalid integer: %s", s)
	}
	return i
}

// FindWorkflows loads all the GitHub Actions workflows for a given repository.
// jwt is the JSON Web Token used for authentication.
// repo is the repository in the format "owner/repo".
func FindWorkflows(client *github.Client, repo string) []string {
	ctx := context.Background()

	// Split repo into owner and name
	parts := strings.Split(strings.Replace(repo, "https://github.com/", "", 1), "/")
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

func WorkflowToString(client *github.Client, repo string, workflow string) string {
	ctx := context.Background()

	// Split repo into owner and name
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return ""
	}
	owner, repoName := parts[0], parts[1]

	// Get file content
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repoName, ".github/workflows/"+workflow, nil)
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

func FindContributorsToWorkflow(client *github.Client, repo string, workflow string) []string {
	ctx := context.Background()

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
func ParseWorkflow(workflow string, workflowId int) []Action {
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

	actionId := 1

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
			actionId++

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
				Id:          fmt.Sprintf("%d-%d", workflowId, actionId),
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

func FindActionsWithDifferentVersions(actions1 []Action, actions2 []Action) ([]Action, int, []string) {
	actions := []Action{}
	count := 0
	result := []string{}

	for _, action1 := range actions1 {
		for _, action2 := range actions2 {

			if action1.Uses != "" && action1.UsesVersion != "" && action2.UsesVersion != "" && action1.Uses == action2.Uses && action1.UsesVersion != action2.UsesVersion {
				if !lo.ContainsBy(actions, func(item Action) bool {
					return item.Id == action1.Id
				}) {
					count += 2
					actions = append(actions, action1)
					actions = append(actions, action2)
				}

				if !slices.Contains(result, action1.Uses) {
					result = append(result, action1.Uses)
				}
			}
		}
	}

	return actions, count, result
}

func FindActionsWithSimilarConfigurations(actions1 []Action, actions2 []Action) ([]Action, int, []string) {

	actions := []Action{}
	result := []string{}
	count := 0

	for _, action1 := range actions1 {
		for _, action2 := range actions2 {
			if action1.Uses == action2.Uses {
				if action1.hash != nil && action2.hash != nil {
					distance := action1.hash.Diff(action2.hash)

					if distance <= HighSimilarity {
						if !lo.ContainsBy(actions, func(item Action) bool {
							return item.Id == action1.Id
						}) {
							count += 2
							actions = append(actions, action1)
							actions = append(actions, action2)
						}

						uses := action1.Uses
						if uses == "" {
							uses = "(script step)"
						}

						if !slices.Contains(result, uses) {
							result = append(result, uses)
						}
					}
				}
			}
		}
	}

	return actions, count, result
}
