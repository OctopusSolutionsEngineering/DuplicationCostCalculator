package workflows

type Report struct {
	NumberOfContributors            int            `json:"numberOfContributors"`
	NumberOfReposWithMixedVersions  int            `json:"numberOfReposWithMixedVersions"`
	NumberOfReposWithSimilarConfigs int            `json:"numberOfReposWithSimilarConfigs"`
	NumberOfReposToBeReviewed       int            `json:"numberOfReposToBeReviewed"`
	SourceReports                   []SourceReport `json:"sourceReports"`
}

type SourceReport struct {
	Repo             string             `json:"repo"`
	RepoMeasurements []RepoMeasurements `json:"repoMeasurements"`
}

type RepoMeasurements struct {
	Repo                          string `json:"repo"`
	StepsWithDifferentVersions    int    `json:"stepsWithDifferentVersions"`
	PercentStepsWithSimilarConfig int    `json:"percentStepsWithSimilarConfig"`
}
