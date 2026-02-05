package workflows

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/collections"
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/models"
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/parsing"
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/infrastructure/githubapi"
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

func GenerateReport(client *github.Client, repos []string) models.Report {
	result := make(chan RepoActions)
	workflowsContent := map[string][]string{}
	workflowsContributors := map[string][]string{}
	repoAdvisories := map[string][]string{}

	for _, repo := range repos {
		// Get the workflows in a goroutine
		go func(client *github.Client, repo string) {

			advisories := githubapi.GetWorkflowAdvisories(client, repo)
			workflowFiles := githubapi.FindWorkflows(client, repo)
			workflows := lo.FilterMap(workflowFiles, func(item string, index int) (string, bool) {
				workflowStr := githubapi.WorkflowToString(client, repo, item)
				return workflowStr, workflowStr != ""
			})
			contributors := lo.Uniq(lo.FlatMap(workflowFiles, func(item string, index int) []string {
				return githubapi.FindContributorsToWorkflow(client, repo, item)
			}))

			// Split repo into owner and name
			owner, repoName, err := parsing.SplitRepo(repo)
			if err != nil {
				owner, repoName = "", ""
			}

			result <- RepoActions{
				Repo:               owner + "/" + repoName,
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

func GenerateReportFromWorkflows(workflows map[string][]string, contributors map[string][]string, repoAdvisories map[string][]string) models.Report {

	repoActions := ConvertWorkflowToActionsMap(workflows)

	repoNames := maps.Keys(repoActions)
	sortedRepoNames := slices.Sorted(repoNames)

	report := models.Report{
		Comparisons:        map[string]map[string]models.RepoMeasurements{},
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

		for j := i + 1; j < len(sortedRepoNames); j++ {
			repo2 := sortedRepoNames[j]
			actionsList2 := repoActions[repo2]

			stepsWithDifferentVersions, diffVersionsIds, stepsWithSimilarConfig, similarConfigIds := GetActionsWithVersionDriftAndDuplication(actionsList1, actionsList2)

			// An overall number of the steps that would have to be updated to ensure consistency between the workflows
			// This includes those that have version drift and those that have similar config
			uniqueActions := lo.Uniq(append(similarConfigIds, diffVersionsIds...))

			if _, ok := report.Comparisons[repo1]; !ok {
				report.Comparisons[repo1] = make(map[string]models.RepoMeasurements)
			}

			if _, ok := report.Comparisons[repo2]; !ok {
				report.Comparisons[repo2] = make(map[string]models.RepoMeasurements)
			}

			report.Comparisons[repo1][repo2] = models.RepoMeasurements{
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

func GetActionsWithVersionDriftAndDuplication(actionsList1 [][]models.Action, actionsList2 [][]models.Action) ([]string, []string, []string, []string) {

	flattenedActionsList1 := lo.Flatten(actionsList1)
	flattenedActionsList2 := lo.Flatten(actionsList2)

	diffVersionsActions, diffVersions := FindActionsWithDifferentVersions(flattenedActionsList1, flattenedActionsList2)
	similarConfigsActions, similarConfigs := FindActionsWithSimilarConfigurations(flattenedActionsList1, flattenedActionsList2)

	// Generate a list of all the action IDs for steps with different versions and similar config
	// This provides a complete list of steps that would have to be updated to ensure consistency between the workflows
	similarConfigIds := lo.UniqMap(similarConfigsActions, func(item models.Action, index int) string {
		return item.Id
	})

	diffVersionsIds := lo.UniqMap(diffVersionsActions, func(item models.Action, index int) string {
		return item.Id
	})

	return diffVersions, diffVersionsIds, similarConfigs, similarConfigIds
}

// CountReposWithDuplicationOrDrift counts the number of repositories that have duplication or drift.
// A repo has duplication or drift if any of its comparisons indicate steps that would need to be updated.
func CountReposWithDuplicationOrDrift(comparisons map[string]map[string]models.RepoMeasurements) int {
	allComparisons := lo.Values(comparisons)
	reposWithDuplicationOrDrift := lo.Filter(allComparisons, func(repoComparisons map[string]models.RepoMeasurements, index int) bool {
		// Get all measurements for this repo's comparisons
		measurements := lo.Values(repoComparisons)
		// Check if any measurement has steps that indicate duplication risk
		measurementsWithRisk := lo.Filter(measurements, func(measurement models.RepoMeasurements, index int) bool {
			return measurement.StepsThatIndicateDuplicationRisk > 0
		})
		// If there's at least one comparison with risk, this repo has duplication or drift
		return len(measurementsWithRisk) > 0
	})
	return len(reposWithDuplicationOrDrift)
}

func ConvertWorkflowToActionsMap(workflows map[string][]string) map[string][][]models.Action {
	repoActions := make(map[string][][]models.Action)

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

// ParseWorkflow parses the string representation of a GitHub Actions workflow
// and returns a slice of Action structs representing the actions used in the workflow.
func ParseWorkflow(workflow string, workflowId int) []models.Action {
	var workflowMap map[string]interface{}

	err := yaml.Unmarshal([]byte(workflow), &workflowMap)
	if err != nil {
		return []models.Action{}
	}

	// Extract jobs
	jobsInterface, ok := workflowMap["jobs"]
	if !ok {
		return []models.Action{}
	}

	jobsMap, ok := jobsInterface.(map[string]interface{})
	if !ok {
		return []models.Action{}
	}

	// 1. Get all keys into a slice
	keys := slices.Collect(maps.Keys(jobsMap))

	// 2. Sort the keys alphabetically
	slices.Sort(keys)

	var actions []models.Action

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
			actionName, actionVersion := parsing.GetActionIdAndVersion(uses)

			// Get the various settings for the action
			env := collections.ConvertStringMap(collections.GetChildMap(stepMap, "env"))
			with := collections.ConvertStringMap(collections.GetChildMap(stepMap, "with"))
			settings := collections.GetOtherValues(stepMap, []string{"uses", "env", "with"})

			action := models.Action{
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

func FindActionsWithDifferentVersions(actions1 []models.Action, actions2 []models.Action) ([]models.Action, []string) {
	actions := []models.Action{}
	result := []string{}

	for _, action1 := range actions1 {
		for _, action2 := range actions2 {

			if HasVersionDrift(action1, action2) {
				if !lo.ContainsBy(actions, func(item models.Action) bool {
					return item.Id == action1.Id
				}) {
					actions = append(actions, action1)
				}

				if !lo.ContainsBy(actions, func(item models.Action) bool {
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

func FindActionsWithSimilarConfigurations(actions1 []models.Action, actions2 []models.Action) ([]models.Action, []string) {

	actions := []models.Action{}
	result := []string{}

	for _, action1 := range actions1 {
		for _, action2 := range actions2 {
			if action1.Uses == action2.Uses {
				if action1.Hash != nil && action2.Hash != nil {
					distance := action1.Hash.Diff(action2.Hash)

					if distance <= HighSimilarity {
						if !lo.ContainsBy(actions, func(item models.Action) bool {
							return item.Id == action1.Id
						}) {
							actions = append(actions, action1)
						}

						if !lo.ContainsBy(actions, func(item models.Action) bool {
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

func GetActionAuthorsFromActionsList(actionsList [][]models.Action) []string {
	if actionsList == nil {
		return []string{}
	}

	return lo.Uniq(lo.FilterMap(lo.Flatten(actionsList), func(item models.Action, index int) (string, bool) {
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
