package workflows

import (
	"fmt"
	"testing"
)

func TestParseWorkflow(t *testing.T) {
	workflowYAML := `
name: CI Pipeline
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'
        env:
          NODE_ENV: production
          
      - name: Run tests
        uses: actions/setup-go@v4.1.0
        with:
          go-version: '1.22'
        env:
          CGO_ENABLED: '0'

      - name: Install Tofu
        run: |
          curl --proto '=https' --tlsv1.2 -fsSL https://get.opentofu.org/install-opentofu.sh -o install-opentofu.sh
          chmod +x install-opentofu.sh
          ./install-opentofu.sh --install-method deb
          rm -f install-opentofu.sh
        shell: bash
          
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`

	actions := ParseWorkflow(workflowYAML, 1)

	if len(actions) != 5 {
		t.Errorf("Expected 5 actions, got %d", len(actions))
	}

	// Test first action
	if actions[0].Uses != "actions/checkout" {
		t.Errorf("Expected first action to be 'actions/checkout', got '%s'", actions[0].Uses)
	}
	if actions[0].UsesVersion != "v4" {
		t.Errorf("Expected first action version to be 'v4', got '%s'", actions[0].UsesVersion)
	}
	if actions[0].With["fetch-depth"] != "0" {
		t.Errorf("Expected fetch-depth to be '0', got '%s'", actions[0].With["fetch-depth"])
	}

	// Test second action with env
	if actions[1].Uses != "actions/setup-node" {
		t.Errorf("Expected second action to be 'actions/setup-node', got '%s'", actions[1].Uses)
	}
	if actions[1].Env["NODE_ENV"] != "production" {
		t.Errorf("Expected NODE_ENV to be 'production', got '%s'", actions[1].Env["NODE_ENV"])
	}
	if actions[1].With["node-version"] != "18" {
		t.Errorf("Expected node-version to be '18', got '%s'", actions[1].With["node-version"])
	}

	// Test third action with version
	if actions[2].Uses != "actions/setup-go" {
		t.Errorf("Expected third action to be 'actions/setup-go', got '%s'", actions[2].Uses)
	}
	if actions[2].UsesVersion != "v4.1.0" {
		t.Errorf("Expected third action version to be 'v4.1.0', got '%s'", actions[2].UsesVersion)
	}

	// Test fourth action (run command, should be ignored)
	if actions[3].Uses != "" {
		t.Errorf("Expected fourth action to be blank, got '%s'", actions[4].Uses)
	}
	if actions[3].Settings["name"] != "Install Tofu" {
		t.Errorf("Expected fourth name to be 'Install Tofu', got '%s'", actions[3].Settings["name"])
	}
	if actions[3].Settings["shell"] != "bash" {
		t.Errorf("Expected fourth shell to be 'bash', got '%s'", actions[3].Settings["shell"])
	}

	// Test fifth action (no version specified)
	if actions[4].Uses != "actions/checkout" {
		t.Errorf("Expected fifth action to be 'actions/checkout', got '%s'", actions[4].Uses)
	}
	if actions[4].UsesVersion != "v3" {
		t.Errorf("Expected fifth action version to be 'v3', got '%s'", actions[4].UsesVersion)
	}

}

func TestParseWorkflowInvalidYAML(t *testing.T) {
	invalidYAML := `
this is not valid yaml: [
`
	actions := ParseWorkflow(invalidYAML, 1)

	if len(actions) != 0 {
		t.Errorf("Expected 0 actions for invalid YAML, got %d", len(actions))
	}
}

func TestParseWorkflowNoJobs(t *testing.T) {
	workflowYAML := `
name: Empty Workflow
on: [push]
`
	actions := ParseWorkflow(workflowYAML, 1)

	if len(actions) != 0 {
		t.Errorf("Expected 0 actions when no jobs defined, got %d", len(actions))
	}
}

func TestParseWorkflowNoSteps(t *testing.T) {
	workflowYAML := `
name: No Steps Workflow
jobs:
  build:
    runs-on: ubuntu-latest
`
	actions := ParseWorkflow(workflowYAML, 1)

	if len(actions) != 0 {
		t.Errorf("Expected 0 actions when no steps defined, got %d", len(actions))
	}
}

func TestFindActionsWithDifferentVersions(t *testing.T) {
	actions1 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v3"},
		{Uses: "actions/setup-node", UsesVersion: "v3"},
		{Uses: "actions/setup-go", UsesVersion: "v4"},
		{Uses: "actions/cache", UsesVersion: "v3"},
	}

	actions2 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v4"},
		{Uses: "actions/setup-node", UsesVersion: "v3"},
		{Uses: "actions/setup-go", UsesVersion: "v5"},
		{Uses: "actions/upload-artifact", UsesVersion: "v3"},
	}

	_, count, _ := FindActionsWithDifferentVersions(actions1, actions2)

	// Expected: checkout (v3 vs v4) and setup-go (v4 vs v5) = 2 differences
	if count != 2 {
		t.Errorf("Expected 2 actions with different versions, got %d", count)
	}
}

func TestFindActionsWithDifferentVersionsNoMatches(t *testing.T) {
	actions1 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v3"},
		{Uses: "actions/setup-node", UsesVersion: "v3"},
	}

	actions2 := []Action{
		{Uses: "actions/setup-go", UsesVersion: "v4"},
		{Uses: "actions/cache", UsesVersion: "v3"},
	}

	_, count, _ := FindActionsWithDifferentVersions(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 actions with different versions, got %d", count)
	}
}

func TestFindActionsWithDifferentVersionsSameVersions(t *testing.T) {
	actions1 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v4"},
		{Uses: "actions/setup-node", UsesVersion: "v3"},
	}

	actions2 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v4"},
		{Uses: "actions/setup-node", UsesVersion: "v3"},
	}

	_, count, _ := FindActionsWithDifferentVersions(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 actions with different versions when all match, got %d", count)
	}
}

func TestFindActionsWithDifferentVersionsEmptyLists(t *testing.T) {
	actions1 := []Action{}
	actions2 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v4"},
	}

	_, count, _ := FindActionsWithDifferentVersions(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 actions with different versions when one list is empty, got %d", count)
	}
}

func TestFindActionsWithDifferentVersionsMultipleSameAction(t *testing.T) {
	actions1 := []Action{
		{Id: "1-1", Uses: "actions/checkout", UsesVersion: "v3"},
		{Id: "1-2", Uses: "actions/checkout", UsesVersion: "v3"},
	}

	actions2 := []Action{
		{Id: "2-1", Uses: "actions/checkout", UsesVersion: "v4"},
		{Id: "2-2", Uses: "actions/checkout", UsesVersion: "v4"},
	}

	_, count, _ := FindActionsWithDifferentVersions(actions1, actions2)

	// Each instance in actions1 matches with each instance in actions2
	// 2 * 2 = 4 differences
	if count != 4 {
		t.Errorf("Expected 4 differences (2x2), got %d", count)
	}
}

func TestFindActionsWithSimilarConfigurations(t *testing.T) {
	actions1 := []Action{
		{
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
				"submodules":  "true",
			},
		},
		{
			Uses:        "actions/setup-node",
			UsesVersion: "v3",
			With: map[string]string{
				"node-version": "18",
				"cache":        "npm",
			},
			Env: map[string]string{
				"NODE_ENV": "production",
			},
		},
	}

	actions2 := []Action{
		{
			Uses:        "actions/checkout",
			UsesVersion: "v3",
			With: map[string]string{
				"fetch-depth": "0",
				"submodules":  "true",
			},
		},
		{
			Uses:        "actions/setup-node",
			UsesVersion: "v3",
			With: map[string]string{
				"node-version": "18",
				"cache":        "npm",
			},
			Env: map[string]string{
				"NODE_ENV": "development",
			},
		},
	}

	// Generate hashes for all actions
	for i := range actions1 {
		actions1[i].GenerateHash()
	}
	for i := range actions2 {
		actions2[i].GenerateHash()
	}

	_, count, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	// Expected: 2 pairs of matching actions with similar configs = 4 total (2 * 2)
	if count < 2 {
		t.Errorf("Expected at least 2 similar configurations, got %d", count)
	}
}

func TestFindActionsWithSimilarConfigurationsNoMatches(t *testing.T) {
	actions1 := []Action{
		{
			Uses:        "actions/checkout",
			UsesVersion: "v4",
		},
	}

	actions2 := []Action{
		{
			Uses:        "actions/setup-node",
			UsesVersion: "v3",
		},
	}

	// Generate hashes
	for i := range actions1 {
		actions1[i].GenerateHash()
	}
	for i := range actions2 {
		actions2[i].GenerateHash()
	}

	_, count, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 similar configurations when action names differ, got %d", count)
	}
}

func TestFindActionsWithSimilarConfigurationsEmptyLists(t *testing.T) {
	actions1 := []Action{}
	actions2 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v4"},
	}

	_, count, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 similar configurations when one list is empty, got %d", count)
	}
}

func TestFindActionsWithSimilarConfigurationsDifferentConfigs(t *testing.T) {
	actions1 := []Action{
		{
			Uses:        "actions/setup-node",
			UsesVersion: "v3",
			With: map[string]string{
				"node-version": "18",
				"cache":        "npm",
				"registry-url": "https://registry.npmjs.org",
			},
			Env: map[string]string{
				"NODE_ENV":     "production",
				"CI":           "true",
				"BUILD_NUMBER": "123",
			},
		},
	}

	actions2 := []Action{
		{
			Uses:        "actions/setup-node",
			UsesVersion: "v3",
			With: map[string]string{
				"node-version": "20",
			},
		},
	}

	// Generate hashes
	for i := range actions1 {
		actions1[i].GenerateHash()
	}
	for i := range actions2 {
		actions2[i].GenerateHash()
	}

	_, count, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	// With significantly different configs, should not count as highly similar
	if count != 0 {
		t.Errorf("Expected 0 for very different configurations, got %d", count)
	}
}

func TestFindActionsWithSimilarConfigurationsNoHashes(t *testing.T) {
	actions1 := []Action{
		{
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
	}

	actions2 := []Action{
		{
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
	}

	// Don't generate hashes - test nil hash handling

	_, count, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 when hashes are not generated, got %d", count)
	}
}

func TestFindActionsWithSimilarConfigurationsMultiplePairs(t *testing.T) {
	actions1 := []Action{
		{
			Id:          "1-1",
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
		{
			Id:          "1-2",
			Uses:        "actions/checkout",
			UsesVersion: "v3",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
	}

	actions2 := []Action{
		{
			Id:          "2-1",
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
		{
			Id:          "2-2",
			Uses:        "actions/checkout",
			UsesVersion: "v3",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
	}

	// Generate hashes
	for i := range actions1 {
		actions1[i].GenerateHash()
	}
	for i := range actions2 {
		actions2[i].GenerateHash()
	}

	_, count, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	// 2 actions in actions1 × 2 actions in actions2 = 4 comparisons
	// All should be highly similar, so count should be 8 (2 per similar pair)
	if count < 4 {
		t.Errorf("Expected at least 4 (2x2x2) similar configurations, got %d", count)
	}
}

func TestGenerateReportFromWorkflows(t *testing.T) {
	workflow1 := `
name: Workflow 1
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
`

	workflow2 := `
name: Workflow 2
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
`

	workflow3 := `
name: Workflow 3
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
`

	workflows := map[string][]string{
		"repo1": {workflow1},
		"repo2": {workflow2},
		"repo3": {workflow3},
	}

	report := GenerateReportFromWorkflows(workflows, map[string][]string{}, map[string][]string{})

	// Verify report structure is initialized
	if report.Comparisons == nil {
		t.Error("Expected Comparisons map to be initialized")
	}

	if report.Comparisons["repo1"]["repo2"].StepsWithDifferentVersionsCount != 2 || len(report.Comparisons["repo1"]["repo2"].StepsWithDifferentVersions) != 1 {
		t.Error("Expected 2 different versions between repo1 and repo2, found " + fmt.Sprintf("%v %v", report.Comparisons["repo1"]["repo2"].StepsWithDifferentVersionsCount, report.Comparisons["repo1"]["repo2"].StepsWithDifferentVersions))
	}

	if report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfigCount != 4 || len(report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfig) != 2 {
		t.Error("Expected 4 similar configurations between repo1 and repo2, found " + fmt.Sprintf("%v", report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfigCount))
	}

	if len(report.Comparisons["repo1"]["repo3"].StepsWithDifferentVersions) != 0 || report.Comparisons["repo1"]["repo3"].StepsWithDifferentVersionsCount != 0 {
		t.Error("Expected 0 different versions between repo1 and repo3, found " + fmt.Sprintf("%v", report.Comparisons["repo1"]["repo3"].StepsWithDifferentVersions))
	}

	if len(report.Comparisons["repo1"]["repo3"].StepsWithSimilarConfig) != 0 || report.Comparisons["repo1"]["repo3"].StepsWithSimilarConfigCount != 0 {
		t.Error("Expected 0 similar configurations between repo1 and repo3, found " + fmt.Sprintf("%v", report.Comparisons["repo1"]["repo3"].StepsWithSimilarConfig))
	}

	if report.Comparisons["repo2"]["repo1"].StepsWithDifferentVersionsCount != 2 || len(report.Comparisons["repo2"]["repo1"].StepsWithDifferentVersions) != 1 {
		t.Error("Expected 2 different versions between repo2 and repo1, found " + fmt.Sprintf("%v %v", report.Comparisons["repo2"]["repo1"].StepsWithDifferentVersionsCount, report.Comparisons["repo2"]["repo1"].StepsWithDifferentVersions))
	}

	if report.Comparisons["repo2"]["repo1"].StepsWithSimilarConfigCount != 4 || len(report.Comparisons["repo2"]["repo1"].StepsWithSimilarConfig) != 2 {
		t.Error("Expected 4 similar configurations between repo2 and repo1, found " + fmt.Sprintf("%v %v", report.Comparisons["repo2"]["repo1"].StepsWithSimilarConfigCount, report.Comparisons["repo2"]["repo1"].StepsWithSimilarConfig))
	}

	if len(report.Comparisons["repo2"]["repo3"].StepsWithDifferentVersions) != 0 || report.Comparisons["repo2"]["repo3"].StepsWithDifferentVersionsCount != 0 {
		t.Error("Expected 0 different versions between repo2 and repo3, found " + fmt.Sprintf("%v", report.Comparisons["repo2"]["repo3"].StepsWithDifferentVersions))
	}

	if len(report.Comparisons["repo2"]["repo3"].StepsWithSimilarConfig) != 0 || report.Comparisons["repo2"]["repo3"].StepsWithSimilarConfigCount != 0 {
		t.Error("Expected 0 similar configurations between repo2 and repo3, found " + fmt.Sprintf("%v", report.Comparisons["repo2"]["repo3"].StepsWithSimilarConfig))
	}

	if len(report.Comparisons["repo3"]["repo1"].StepsWithDifferentVersions) != 0 || report.Comparisons["repo3"]["repo1"].StepsWithDifferentVersionsCount != 0 {
		t.Error("Expected 0 different versions between repo3 and repo1, found " + fmt.Sprintf("%v", report.Comparisons["repo3"]["repo1"].StepsWithDifferentVersions))
	}

	if len(report.Comparisons["repo3"]["repo1"].StepsWithSimilarConfig) != 0 || report.Comparisons["repo3"]["repo1"].StepsWithSimilarConfigCount != 0 {
		t.Error("Expected 0 similar configurations between repo3 and repo1, found " + fmt.Sprintf("%v", report.Comparisons["repo3"]["repo1"].StepsWithSimilarConfig))
	}

	if len(report.Comparisons["repo3"]["repo2"].StepsWithDifferentVersions) != 0 || report.Comparisons["repo3"]["repo2"].StepsWithDifferentVersionsCount != 0 {
		t.Error("Expected 0 different versions between repo3 and repo2, found " + fmt.Sprintf("%v", report.Comparisons["repo3"]["repo2"].StepsWithDifferentVersions))
	}

	if len(report.Comparisons["repo3"]["repo2"].StepsWithSimilarConfig) != 0 || report.Comparisons["repo3"]["repo2"].StepsWithSimilarConfigCount != 0 {
		t.Error("Expected 0 similar configurations between repo3 and repo2, found " + fmt.Sprintf("%v", report.Comparisons["repo3"]["repo2"].StepsWithSimilarConfig))
	}
}

func TestGenerateReportFromWorkflowsEmpty(t *testing.T) {
	workflows := map[string][]string{}

	report := GenerateReportFromWorkflows(workflows, map[string][]string{}, map[string][]string{})

	if report.Comparisons == nil {
		t.Error("Expected Comparisons map to be initialized even for empty input")
	}
}

func TestGenerateReportFromWorkflowsSingleRepo(t *testing.T) {
	workflow := `
name: Single Workflow
jobs:
  build:
    steps:
      - uses: actions/checkout@v3
`

	workflows := map[string][]string{
		"repo1": {workflow},
	}

	report := GenerateReportFromWorkflows(workflows, map[string][]string{}, map[string][]string{})

	// With only one repo, there should be no comparisons
	if report.Comparisons == nil {
		t.Error("Expected Comparisons map to be initialized")
	}

	if len(report.Comparisons) != 0 {
		t.Error("Expected Comparisons map to be initialized even for single input")
	}
}

func TestGenerateReportFromWorkflowsDissimilarSteps(t *testing.T) {
	workflow := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Script 1
        run: This is a test script
`

	workflow2 := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Some script step
        run: These scripts are nothing alike
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

	if report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfigCount != 0 || len(report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfig) != 0 {
		t.Error("Expected no similar steps, found " + fmt.Sprintf("%v %v", report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfigCount, report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfig))
	}
}

func TestGenerateReportFromWorkflowsDissimilarSteps2(t *testing.T) {
	workflow := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Download JUnit Summary from Previous Workflow
        id: download-artifact
        uses: dawidd6/action-download-artifact@v6
        with:
          workflow_conclusion: success
          name: junit-test-summary
          if_no_artifact_found: warn
          branch: main
`

	workflow2 := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Download JUnit Summary from Previous Workflow
        id: download-artifact
        uses: dawidd6/action-download-artifact@v6
        with:
          workflow_conclusion: failure
          name: a-different-name
          if_no_artifact_found: error
          branch: feature
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

	if report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfigCount != 0 || len(report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfig) != 0 {
		t.Error("Expected no similar steps, found " + fmt.Sprintf("%v %v", report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfigCount, report.Comparisons["repo1"]["repo2"].StepsWithSimilarConfig))
	}
}

func TestGenerateReportFromWorkflowsSimilarSteps(t *testing.T) {
	workflow := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Script 1
        run: This is a test script
`

	workflow2 := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Script 1
        run: This is a testing script
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

func TestGenerateReportFromWorkflowsSimilarSteps2(t *testing.T) {
	workflow := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Download JUnit Summary from Previous Workflow
        id: download-artifact
        uses: dawidd6/action-download-artifact@v6
        with:
          workflow_conclusion: success
          name: junit-test-summary
          if_no_artifact_found: warn
          branch: main
`

	workflow2 := `
name: Single Workflow
jobs:
  build:
    steps:
      - name: Download JUnit Summary from Previous Workflow
        id: download-artifact
        uses: dawidd6/action-download-artifact@v6
        with:
          workflow_conclusion: success
          name: junit-test-summary-2
          if_no_artifact_found: warn
          branch: main2
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
