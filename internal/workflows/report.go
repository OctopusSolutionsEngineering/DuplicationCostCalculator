package workflows

type Report struct {
	NumberOfContributors            int                                    `json:"numberOfContributors"`
	NumberOfReposWithMixedVersions  int                                    `json:"numberOfReposWithMixedVersions"`
	NumberOfReposWithSimilarConfigs int                                    `json:"numberOfReposWithSimilarConfigs"`
	NumberOfReposToBeReviewed       int                                    `json:"numberOfReposToBeReviewed"`
	Comparisons                     map[string]map[string]RepoMeasurements `json:"comparisons"`
}

type RepoMeasurements struct {
	StepsWithDifferentVersions int `json:"stepsWithDifferentVersions"`
	StepsWithSimilarConfig     int `json:"stepsWithSimilarConfig"`
}
