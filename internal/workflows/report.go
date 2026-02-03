package workflows

type Report struct {
	NumberOfRepos                       int                                    `json:"numberOfRepos"`
	NumberOfReposWithDuplicationOrDrift int                                    `json:"numberOfReposWithDuplicationOrDrift"`
	Comparisons                         map[string]map[string]RepoMeasurements `json:"comparisons"`
	Contributors                        map[string][]string                    `json:"contributors"`
	UniqueContributors                  []string                               `json:"uniqueContributors"`
	WorkflowAdvisories                  map[string][]string                    `json:"workflowAdvisories"`
}

type RepoMeasurements struct {
	StepsWithDifferentVersions       []string `json:"stepsWithDifferentVersions"`
	StepsWithDifferentVersionsCount  int      `json:"stepsWithDifferentVersionsCount"`
	StepsWithSimilarConfig           []string `json:"stepsWithSimilarConfig"`
	StepsWithSimilarConfigCount      int      `json:"stepsWithSimilarConfigCount"`
	StepsThatIndicateDuplicationRisk int      `json:"stepsThatIndicateDuplicationRisk"`
}
