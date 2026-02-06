package workflows

import (
	"fmt"
	"testing"
)

func TestGenerateReportFromWorkflowsSimilarScripts(t *testing.T) {
	workflow := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Script 1
        run: |
          Echo "This is a test"
          Echo "This is a test2"
          Echo "This is a test3"
          Echo "This is a test4"
          Echo "This is a test5"
          Echo "This is a test6"
          Echo "This is a test7"
          Echo "This is a test8"
          Echo "This is a test9"
`

	workflow2 := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Script 1
        run: |
          Echo "This is a testa"
          Echo "This is a test2"
          Echo "This is a test3"
          Echo "This is a test4"
          Echo "This is a test5"
          Echo "This is a test6"
          Echo "This is a test7"
          Echo "This is a test8"
          Echo "This is a test10"
`

	workflows := map[string][]string{
		"repo1": {workflow},
		"repo2": {workflow2},
	}

	report := GenerateReportFromWorkflows(workflows, map[string][]string{}, map[string][]string{})

	// With only one repo, there should be no comparisons
	if report.Comparisons == nil {
		t.Error("Expected Comparisons map to be initialized")
	}

	if report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfigCount != 2 || len(report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfig) != 1 {
		t.Error("Expected one similar steps, found " + fmt.Sprintf("%v %v", report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfigCount, report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfig))
	}
}
