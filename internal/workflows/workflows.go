package workflows

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/collections"
	"github.com/google/go-github/v57/github"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

const HighSimilarity = 30
const BuiltInStep = "(built-in step)"

type RepoActions struct {
	Repo               string
	Workflows          []string
	Contributors       []string
	WorkflowAdvisories []string
}

func GenerateReport(client *github.Client, repos []string) Report {
	result := make(chan RepoActions)
	workflowsContent := map[string][]string{}
	workflowsContributors := map[string][]string{}
	repoAdvisories := map[string][]string{}

	for _, repo := range repos {
		// Get the workflows in a goroutine
		go func(client *github.Client, repo string) {
			workflows := []string{}
			contributors := []string{}
			advisories := GetWorkflowAdvisories(client, repo)
			workflowFiles := FindWorkflows(client, repo)

			for _, workflowFile := range workflowFiles {
				workflowStr := WorkflowToString(client, repo, workflowFile)
				if workflowStr != "" {
					workflows = append(workflows, workflowStr)
				}
				contributors = FindContributorsToWorkflow(client, repo, workflowFile)
			}

			result <- RepoActions{
				Repo:               repo,
				Workflows:          workflows,
				Contributors:       contributors,
				WorkflowAdvisories: advisories,
			}
		}(client, repo)
	}

	// Wait for all the goroutines to finish
	for i := 0; i < len(repos); i++ {
		repoActions := <-result
		workflowsContent[repoActions.Repo] = repoActions.Workflows
		workflowsContributors[repoActions.Repo] = repoActions.Contributors
		repoAdvisories[repoActions.Repo] = repoActions.WorkflowAdvisories
	}

	report := GenerateReportFromWorkflows(workflowsContent, workflowsContributors, repoAdvisories)

	return report
}

func GenerateReportFromWorkflows(workflows map[string][]string, contributors map[string][]string, repoAdvisories map[string][]string) Report {

	repoActions := convertWorkflowToActionsMap(workflows)

	repoNames := maps.Keys(repoActions)
	sortedRepoNames := slices.Sorted(repoNames)

	report := Report{
		Comparisons:        map[string]map[string]RepoMeasurements{},
		Contributors:       map[string][]string{},
		WorkflowAdvisories: map[string][]string{},
		ActionAuthors:      map[string][]string{},
		NumberOfRepos:      len(sortedRepoNames),
	}

	for i := 0; i < len(sortedRepoNames); i++ {
		repo1 := sortedRepoNames[i]
		actionsList1 := repoActions[repo1]
		report.Contributors[repo1] = contributors[repo1]
		report.WorkflowAdvisories[repo1] = repoAdvisories[repo1]
		report.ActionAuthors[repo1] = GetActionAuthorsFromActionsList(actionsList1)

		uniqueActions := []string{}

		for j := i + 1; j < len(sortedRepoNames); j++ {
			repo2 := sortedRepoNames[j]
			actionsList2 := repoActions[repo2]

			// The unique list of uses values that identity the kinds of steps that have version drift
			stepsWithDifferentVersions := []string{}
			// The unique list of uses values that identity the kinds of steps that have similar config
			stepsWithSimilarConfig := []string{}

			diffVersionsIds := []string{}
			similarConfigIds := []string{}

			for _, actions1 := range actionsList1 {
				for _, actions2 := range actionsList2 {
					diffVersionsActions, diffVersions := FindActionsWithDifferentVersions(actions1, actions2)
					similarConfigsActions, similarConfigs := FindActionsWithSimilarConfigurations(actions1, actions2)

					// Generate a list of all the "uses" values for steps with different versions and similar config, ensuring uniqueness
					// This provides a quick way to idtenify the kinds of steps that are contributing to duplication or drift
					stepsWithDifferentVersions = lo.Uniq(append(stepsWithDifferentVersions, diffVersions...))
					stepsWithSimilarConfig = lo.Uniq(append(stepsWithSimilarConfig, similarConfigs...))

					// Generate a list of all the action IDs for steps with different versions and similar config
					// This provides a complete list of steps that would have to be updated to ensure consistency between the workflows
					similarConfigIds = lo.Uniq(append(similarConfigIds, lo.Map(similarConfigsActions, func(item Action, index int) string {
						return item.Id
					})...))

					diffVersionsIds = lo.Uniq(append(diffVersionsIds, lo.Map(diffVersionsActions, func(item Action, index int) string {
						return item.Id
					})...))

					// An overall number of the steps that would have to be updated to ensure consistency between the workflows
					// This includes those that have version drift and those that have similar config
					uniqueActions = lo.Uniq(append(similarConfigIds, diffVersionsIds...))
				}
			}

			if _, ok := report.Comparisons[repo1]; !ok {
				report.Comparisons[repo1] = make(map[string]RepoMeasurements)
			}

			if _, ok := report.Comparisons[repo2]; !ok {
				report.Comparisons[repo2] = make(map[string]RepoMeasurements)
			}

			report.Comparisons[repo1][repo2] = RepoMeasurements{
				StepsWithDifferentVersions:       stepsWithDifferentVersions,
				StepsWithDifferentVersionsCount:  len(diffVersionsIds),
				StepsWithSimilarConfig:           stepsWithSimilarConfig,
				StepsWithSimilarConfigCount:      len(similarConfigIds),
				StepsThatIndicateDuplicationRisk: len(uniqueActions),
			}

			// The measurements for repo2 compared to repo1 are the same as repo1 compared to repo2,
			// so we can copy them over instead of recalculating
			report.Comparisons[repo2][repo1] = report.Comparisons[repo1][repo2]
		}
	}

	// Count the number of repositories that have duplication or drift
	report.NumberOfReposWithDuplicationOrDrift = CountReposWithDuplicationOrDrift(report.Comparisons)

	// Get all unique contributors across all repositories
	allContributorLists := lo.Values(report.Contributors)
	flattenedContributors := lo.Flatten(allContributorLists)
	report.UniqueContributors = lo.Uniq(flattenedContributors)

	return report
}

// CountReposWithDuplicationOrDrift counts the number of repositories that have duplication or drift.
// A repo has duplication or drift if any of its comparisons indicate steps that would need to be updated.
func CountReposWithDuplicationOrDrift(comparisons map[string]map[string]RepoMeasurements) int {
	allComparisons := lo.Values(comparisons)
	reposWithDuplicationOrDrift := lo.Filter(allComparisons, func(repoComparisons map[string]RepoMeasurements, index int) bool {
		// Get all measurements for this repo's comparisons
		measurements := lo.Values(repoComparisons)
		// Check if any measurement has steps that indicate duplication risk
		measurementsWithRisk := lo.Filter(measurements, func(measurement RepoMeasurements, index int) bool {
			return measurement.StepsThatIndicateDuplicationRisk > 0
		})
		// If there's at least one comparison with risk, this repo has duplication or drift
		return len(measurementsWithRisk) > 0
	})
	return len(reposWithDuplicationOrDrift)
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

// FindWorkflows loads all the GitHub Actions workflows for a given repository.
// jwt is the JSON Web Token used for authentication.
// repo is the repository in the format "owner/repo".
func FindWorkflows(client *github.Client, repo string) []string {
	ctx := context.Background()

	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return []string{}
	}

	// List contents of .github/workflows directory
	_, dirContent, _, err := client.Repositories.GetContents(ctx, owner, repoName, ".github/workflows", nil)
	if err != nil {
		println("Error fetching workflows for repo", repo, ":", err.Error())
		return []string{}
	}

	files := lo.Filter(dirContent, func(item *github.RepositoryContent, index int) bool {
		return item.GetType() == "file" && strings.HasSuffix(strings.ToLower(item.GetName()), ".yml") || strings.HasSuffix(strings.ToLower(item.GetName()), ".yaml")
	})

	return lo.Map(files, func(item *github.RepositoryContent, index int) string {
		return item.GetName()
	})
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
	var contributors []string

	// Fetch all commits for the workflow file (handle pagination)
	for {
		commits, resp, err := client.Repositories.ListCommits(ctx, owner, repoName, opts)
		if err != nil {
			return []string{}
		}

		// Extract unique contributor names
		authors := lo.Filter(commits, func(item *github.RepositoryCommit, index int) bool {
			return item.Commit != nil && item.Commit.Author != nil && item.Commit.Author.Name != nil
		})

		authorNames := lo.Map(authors, func(item *github.RepositoryCommit, index int) string {
			return *item.Commit.Author.Name
		})

		contributors = lo.Uniq(append(contributors, authorNames...))

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

			uses := collections.GetStringProperty(stepMap, "uses")

			// Split uses into action and version
			actionName, actionVersion := GetActionIdAndVersion(uses)

			// Get the various settings for the action
			env := collections.ConvertStringMap(collections.GetChildMap(stepMap, "env"))
			with := collections.ConvertStringMap(collections.GetChildMap(stepMap, "with"))
			settings := collections.GetOtherValues(stepMap, []string{"uses", "env", "with"})

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

func FindActionsWithDifferentVersions(actions1 []Action, actions2 []Action) ([]Action, []string) {
	actions := []Action{}
	result := []string{}

	for _, action1 := range actions1 {
		for _, action2 := range actions2 {

			if HasVersionDrift(action1, action2) {
				if !lo.ContainsBy(actions, func(item Action) bool {
					return item.Id == action1.Id
				}) {
					actions = append(actions, action1)
				}

				if !lo.ContainsBy(actions, func(item Action) bool {
					return item.Id == action2.Id
				}) {
					actions = append(actions, action2)
				}

				if !slices.Contains(result, action1.Uses) {
					result = append(result, action1.Uses)
				}
			}
		}
	}

	return actions, result
}

func FindActionsWithSimilarConfigurations(actions1 []Action, actions2 []Action) ([]Action, []string) {

	actions := []Action{}
	result := []string{}

	for _, action1 := range actions1 {
		for _, action2 := range actions2 {
			if action1.Uses == action2.Uses {
				if action1.hash != nil && action2.hash != nil {
					distance := action1.hash.Diff(action2.hash)

					if distance <= HighSimilarity {
						if !lo.ContainsBy(actions, func(item Action) bool {
							return item.Id == action1.Id
						}) {
							actions = append(actions, action1)
						}

						if !lo.ContainsBy(actions, func(item Action) bool {
							return item.Id == action2.Id
						}) {
							actions = append(actions, action2)
						}

						uses := action1.Uses
						if uses == "" {
							uses = BuiltInStep
						}

						if !slices.Contains(result, uses) {
							result = append(result, uses)
						}
					}
				}
			}
		}
	}

	return actions, result
}

func GetWorkflowAdvisories(client *github.Client, repo string) []string {
	ctx := context.Background()

	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return []string{}
	}

	opts := &github.ListRepositorySecurityAdvisoriesOptions{
		ListCursorOptions: github.ListCursorOptions{
			PerPage: 100,
		},
	}

	var advisories []string

	// Fetch all security advisories for the repository (handle pagination)
	for {
		advisoryList, resp, err := client.SecurityAdvisories.ListRepositorySecurityAdvisories(ctx, owner, repoName, opts)
		if err != nil {
			// If there's an error (e.g., no access, repo doesn't exist), return empty list
			return []string{}
		}

		// Extract advisory IDs or summaries
		for _, advisory := range advisoryList {
			if advisory.GHSAID != nil {
				advisories = append(advisories, *advisory.GHSAID)
			}
		}

		// Check if there are more pages
		if resp.After == "" {
			break
		}
		opts.After = resp.After
	}

	if advisories == nil {
		return []string{}
	}

	return advisories
}

func splitRepo(repo string) (string, string, error) {
	// Split repo into owner and name
	parts := strings.Split(strings.Replace(repo, "https://github.com/", "", 1), "/")
	// Ignore any paths that may have been on the end of a url
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository format: %s", repo)
	}
	return parts[0], parts[1], nil
}

func GetActionAuthorsFromActionsList(actionsList [][]Action) []string {
	if actionsList == nil {
		return []string{}
	}

	return lo.Uniq(lo.FilterMap(lo.Flatten(actionsList), func(item Action, index int) (string, bool) {
		split := strings.Split(item.Uses, "/")
		if len(split) > 0 {
			if split[0] == "" {
				return BuiltInStep, true
			}
			return split[0], true
		}
		return "", false
	}))
}
