package workflows

type Report struct {
	NumberOfContributors            int                                    `json:"numberOfContributors"`
	NumberOfReposWithMixedVersions  int                                    `json:"numberOfReposWithMixedVersions"`
	NumberOfReposWithSimilarConfigs int                                    `json:"numberOfReposWithSimilarConfigs"`
	NumberOfReposToBeReviewed       int                                    `json:"numberOfReposToBeReviewed"`
	Comparisons                     map[string]map[string]RepoMeasurements `json:"comparisons"`
}

type RepoMeasurements struct {
	StepsWithDifferentVersions       []string `json:"stepsWithDifferentVersions"`
	StepsWithDifferentVersionsCount  int      `json:"stepsWithDifferentVersionsCount"`
	StepsWithSimilarConfig           []string `json:"stepsWithSimilarConfig"`
	StepsWithSimilarConfigCount      int      `json:"stepsWithSimilarConfigCount"`
	StepsThatIndicateDuplicationRisk int      `json:"stepsThatIndicateDuplicationRisk"`
}
