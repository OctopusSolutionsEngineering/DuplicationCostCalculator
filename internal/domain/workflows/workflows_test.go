package workflows

import (
	"fmt"
	"testing"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/models"
)

func TestCountReposWithDuplicationOrDrift(t *testing.T) {
	tests := []struct {
		name        string
		comparisons map[string]map[string]models.RepoMeasurements
		expected    int
	}{
		{
			name:        "empty comparisons",
			comparisons: map[string]map[string]models.RepoMeasurements{},
			expected:    0,
		},
		{
			name: "no duplication or drift",
			comparisons: map[string]map[string]models.RepoMeasurements{
				"repo1": {
					"repo2": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 0,
					},
				},
				"repo2": {
					"repo1": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 0,
					},
				},
			},
			expected: 0,
		},
		{
			name: "one repo with duplication",
			comparisons: map[string]map[string]models.RepoMeasurements{
				"repo1": {
					"repo2": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 5,
					},
				},
				"repo2": {
					"repo1": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 5,
					},
				},
			},
			expected: 2,
		},
		{
			name: "multiple repos with duplication",
			comparisons: map[string]map[string]models.RepoMeasurements{
				"repo1": {
					"repo2": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 3,
					},
					"repo3": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 2,
					},
				},
				"repo2": {
					"repo1": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 3,
					},
					"repo3": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 1,
					},
				},
				"repo3": {
					"repo1": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 2,
					},
					"repo2": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 1,
					},
				},
			},
			expected: 3,
		},
		{
			name: "mixed - some with duplication, some without",
			comparisons: map[string]map[string]models.RepoMeasurements{
				"repo1": {
					"repo2": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 5,
					},
					"repo3": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 0,
					},
				},
				"repo2": {
					"repo1": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 5,
					},
					"repo3": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 0,
					},
				},
				"repo3": {
					"repo1": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 0,
					},
					"repo2": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 0,
					},
				},
			},
			expected: 2,
		},
		{
			name: "single comparison with duplication",
			comparisons: map[string]map[string]models.RepoMeasurements{
				"repo1": {
					"repo2": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 10,
					},
				},
			},
			expected: 1,
		},
		{
			name: "repo with multiple comparisons, only one has duplication",
			comparisons: map[string]map[string]models.RepoMeasurements{
				"repo1": {
					"repo2": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 0,
					},
					"repo3": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 0,
					},
					"repo4": models.RepoMeasurements{
						StepsThatIndicateDuplicationRisk: 1,
					},
				},
			},
			expected: 1,
		},
		{
			name: "large number of repos with varying duplication",
			comparisons: map[string]map[string]models.RepoMeasurements{
				"repo1": {
					"repo2": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 10},
					"repo3": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 5},
					"repo4": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
					"repo5": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 8},
				},
				"repo2": {
					"repo1": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 10},
					"repo3": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 3},
					"repo4": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
					"repo5": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
				},
				"repo3": {
					"repo1": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 5},
					"repo2": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 3},
					"repo4": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
					"repo5": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
				},
				"repo4": {
					"repo1": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
					"repo2": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
					"repo3": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
					"repo5": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
				},
				"repo5": {
					"repo1": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 8},
					"repo2": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
					"repo3": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
					"repo4": models.RepoMeasurements{StepsThatIndicateDuplicationRisk: 0},
				},
			},
			expected: 4, // repo1, repo2, repo3, repo5 all have at least one comparison with duplication
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountReposWithDuplicationOrDrift(tt.comparisons)
			if result != tt.expected {
				t.Errorf("CountReposWithDuplicationOrDrift() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestConvertWorkflowToActionsMap(t *testing.T) {
	tests := []struct {
		name                 string
		workflows            map[string][]string
		expectedRepoCount    int
		expectedWorkflowsFor map[string]int // repo -> number of workflows
	}{
		{
			name:              "empty workflows map",
			workflows:         map[string][]string{},
			expectedRepoCount: 0,
		},
		{
			name: "single repo with single workflow",
			workflows: map[string][]string{
				"owner/repo1": {
					`
name: Test Workflow
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
				},
			},
			expectedRepoCount: 1,
			expectedWorkflowsFor: map[string]int{
				"owner/repo1": 1,
			},
		},
		{
			name: "single repo with multiple workflows",
			workflows: map[string][]string{
				"owner/repo1": {
					`
name: Workflow 1
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
					`
name: Workflow 2
on: [pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v3
`,
				},
			},
			expectedRepoCount: 1,
			expectedWorkflowsFor: map[string]int{
				"owner/repo1": 2,
			},
		},
		{
			name: "multiple repos with single workflow each",
			workflows: map[string][]string{
				"owner/repo1": {
					`
name: Test Workflow
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
				},
				"owner/repo2": {
					`
name: Build Workflow
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v4
`,
				},
			},
			expectedRepoCount: 2,
			expectedWorkflowsFor: map[string]int{
				"owner/repo1": 1,
				"owner/repo2": 1,
			},
		},
		{
			name: "multiple repos with multiple workflows",
			workflows: map[string][]string{
				"owner/repo1": {
					`
name: Workflow 1
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
					`
name: Workflow 2
on: [pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v3
`,
				},
				"owner/repo2": {
					`
name: CI
on: [push]
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
`,
				},
			},
			expectedRepoCount: 2,
			expectedWorkflowsFor: map[string]int{
				"owner/repo1": 2,
				"owner/repo2": 1,
			},
		},
		{
			name: "workflow with no steps",
			workflows: map[string][]string{
				"owner/repo1": {
					`
name: Empty Workflow
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
`,
				},
			},
			expectedRepoCount: 1,
			expectedWorkflowsFor: map[string]int{
				"owner/repo1": 1,
			},
		},
		{
			name: "invalid workflow YAML",
			workflows: map[string][]string{
				"owner/repo1": {
					`this is not valid yaml: [[[`,
				},
			},
			expectedRepoCount: 1,
			expectedWorkflowsFor: map[string]int{
				"owner/repo1": 1,
			},
		},
		{
			name: "workflow with multiple jobs and steps",
			workflows: map[string][]string{
				"owner/repo1": {
					`
name: Complex Workflow
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v5
`,
				},
			},
			expectedRepoCount: 1,
			expectedWorkflowsFor: map[string]int{
				"owner/repo1": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertWorkflowToActionsMap(tt.workflows)

			// Check number of repos
			if len(result) != tt.expectedRepoCount {
				t.Errorf("ConvertWorkflowToActionsMap() returned %d repos, expected %d", len(result), tt.expectedRepoCount)
			}

			// Check number of workflows per repo
			for repo, expectedCount := range tt.expectedWorkflowsFor {
				if actions, ok := result[repo]; !ok {
					t.Errorf("ConvertWorkflowToActionsMap() missing repo %q", repo)
				} else if len(actions) != expectedCount {
					t.Errorf("ConvertWorkflowToActionsMap()[%q] has %d workflows, expected %d", repo, len(actions), expectedCount)
				}
			}
		})
	}
}

func TestConvertWorkflowToActionsMapWorkflowIdUniqueness(t *testing.T) {
	// Test that workflow IDs are unique across all workflows
	workflows := map[string][]string{
		"owner/repo1": {
			`
name: Workflow 1
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
			`
name: Workflow 2
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v3
`,
		},
		"owner/repo2": {
			`
name: Workflow 3
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v4
`,
		},
	}

	result := ConvertWorkflowToActionsMap(workflows)

	// Collect all action IDs
	seenIds := make(map[string]bool)
	for _, repoWorkflows := range result {
		for _, workflow := range repoWorkflows {
			for _, action := range workflow {
				if seenIds[action.Id] {
					t.Errorf("ConvertWorkflowToActionsMap() generated duplicate action ID: %q", action.Id)
				}
				seenIds[action.Id] = true
			}
		}
	}
}

func TestConvertWorkflowToActionsMapPreservesOrder(t *testing.T) {
	// Test that workflows are processed in order (though map iteration is random,
	// the function should handle all workflows)
	workflows := map[string][]string{
		"owner/repo1": {
			`
name: First Workflow
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
			`
name: Second Workflow
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v3
`,
			`
name: Third Workflow
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v4
`,
		},
	}

	result := ConvertWorkflowToActionsMap(workflows)

	// Verify all three workflows were processed
	if len(result["owner/repo1"]) != 3 {
		t.Errorf("ConvertWorkflowToActionsMap() processed %d workflows, expected 3", len(result["owner/repo1"]))
	}
}

func TestConvertWorkflowToActionsMapEmptyWorkflowStrings(t *testing.T) {
	// Test handling of empty workflow strings
	workflows := map[string][]string{
		"owner/repo1": {
			"",
			`
name: Valid Workflow
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`,
			"",
		},
	}

	result := ConvertWorkflowToActionsMap(workflows)

	// Should process all three "workflows" even if some are empty/invalid
	if len(result["owner/repo1"]) != 3 {
		t.Errorf("ConvertWorkflowToActionsMap() processed %d workflows, expected 3", len(result["owner/repo1"]))
	}
}

func TestConvertWorkflowToActionsMapNilWorkflows(t *testing.T) {
	// Test with nil workflows map
	var workflows map[string][]string = nil

	result := ConvertWorkflowToActionsMap(workflows)

	if result == nil {
		t.Error("ConvertWorkflowToActionsMap() returned nil, expected empty map")
	}

	if len(result) != 0 {
		t.Errorf("ConvertWorkflowToActionsMap() returned %d repos, expected 0", len(result))
	}
}

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
	actions1 := []models.Action{
		{Id: "1-1", Uses: "actions/checkout", UsesVersion: "v3"},
		{Id: "1-2", Uses: "actions/setup-node", UsesVersion: "v3"},
		{Id: "1-3", Uses: "actions/setup-go", UsesVersion: "v4"},
		{Id: "1-4", Uses: "actions/cache", UsesVersion: "v3"},
	}

	actions2 := []models.Action{
		{Id: "2-1", Uses: "actions/checkout", UsesVersion: "v4"},
		{Id: "2-2", Uses: "actions/setup-node", UsesVersion: "v3"},
		{Id: "2-3", Uses: "actions/setup-go", UsesVersion: "v5"},
		{Id: "2-4", Uses: "actions/upload-artifact", UsesVersion: "v3"},
	}

	items, names := FindActionsWithDifferentVersions(actions1, actions2)

	// Expected: checkout (v3 vs v4) and setup-go (v4 vs v5) = 2 differences
	if len(names) != 2 {
		t.Errorf("Expected 2 action types with different versions, got %d %v", len(names), names)
	}

	if len(items) != 4 {
		t.Errorf("Expected 4 actions with different versions, got %d %v", len(items), items)
	}
}

func TestFindActionsWithDifferentVersionsNoMatches(t *testing.T) {
	actions1 := []models.Action{
		{Id: "1-1", Uses: "actions/checkout", UsesVersion: "v3"},
		{Id: "1-2", Uses: "actions/setup-node", UsesVersion: "v3"},
	}

	actions2 := []models.Action{
		{Id: "2-1", Uses: "actions/setup-go", UsesVersion: "v4"},
		{Id: "2-2", Uses: "actions/cache", UsesVersion: "v3"},
	}

	items, _ := FindActionsWithDifferentVersions(actions1, actions2)

	if len(items) != 0 {
		t.Errorf("Expected 0 actions with different versions, got %d", len(items))
	}
}

func TestFindActionsWithDifferentVersionsSameVersions(t *testing.T) {
	actions1 := []models.Action{
		{Id: "1-1", Uses: "actions/checkout", UsesVersion: "v4"},
		{Id: "1-2", Uses: "actions/setup-node", UsesVersion: "v3"},
	}

	actions2 := []models.Action{
		{Id: "2-1", Uses: "actions/checkout", UsesVersion: "v4"},
		{Id: "2-2", Uses: "actions/setup-node", UsesVersion: "v3"},
	}

	items, _ := FindActionsWithDifferentVersions(actions1, actions2)

	if len(items) != 0 {
		t.Errorf("Expected 0 actions with different versions when all match, got %d", len(items))
	}
}

func TestFindActionsWithDifferentVersionsEmptyLists(t *testing.T) {
	actions1 := []models.Action{}
	actions2 := []models.Action{
		{Id: "2-1", Uses: "actions/checkout", UsesVersion: "v4"},
	}

	items, _ := FindActionsWithDifferentVersions(actions1, actions2)

	if len(items) != 0 {
		t.Errorf("Expected 0 actions with different versions when one list is empty, got %d", len(items))
	}
}

func TestFindActionsWithDifferentVersionsMultipleSameAction(t *testing.T) {
	actions1 := []models.Action{
		{Id: "1-1", Uses: "actions/checkout", UsesVersion: "v3"},
		{Id: "1-2", Uses: "actions/checkout", UsesVersion: "v3"},
	}

	actions2 := []models.Action{
		{Id: "2-1", Uses: "actions/checkout", UsesVersion: "v4"},
		{Id: "2-2", Uses: "actions/checkout", UsesVersion: "v4"},
	}

	items, _ := FindActionsWithDifferentVersions(actions1, actions2)

	// Each instance in actions1 matches with each instance in actions2
	// 2 * 2 = 4 differences
	if len(items) != 4 {
		t.Errorf("Expected 4 differences (2x2), got %d", len(items))
	}
}

func TestFindActionsWithSimilarConfigurations(t *testing.T) {
	actions1 := []models.Action{
		{
			Id:          "1-1",
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
				"submodules":  "true",
			},
		},
		{
			Id:          "1-2",
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

	actions2 := []models.Action{
		{
			Id:          "2-1",
			Uses:        "actions/checkout",
			UsesVersion: "v3",
			With: map[string]string{
				"fetch-depth": "0",
				"submodules":  "true",
			},
		},
		{
			Id:          "2-2",
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

	items, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	// Expected: 2 pairs of matching actions with similar configs = 4 total (2 * 2)
	if len(items) < 2 {
		t.Errorf("Expected at least 2 similar configurations, got %d", len(items))
	}
}

func TestFindActionsWithSimilarConfigurationsNoMatches(t *testing.T) {
	actions1 := []models.Action{
		{
			Id:          "1-1",
			Uses:        "actions/checkout",
			UsesVersion: "v4",
		},
	}

	actions2 := []models.Action{
		{
			Id:          "2-1",
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

	items, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	if len(items) != 0 {
		t.Errorf("Expected 0 similar configurations when action names differ, got %d", len(items))
	}
}

func TestFindActionsWithSimilarConfigurationsEmptyLists(t *testing.T) {
	actions1 := []models.Action{}
	actions2 := []models.Action{
		{Id: "2-1", Uses: "actions/checkout", UsesVersion: "v4"},
	}

	items, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	if len(items) != 0 {
		t.Errorf("Expected 0 similar configurations when one list is empty, got %d", len(items))
	}
}

func TestFindActionsWithSimilarConfigurationsDifferentConfigs(t *testing.T) {
	actions1 := []models.Action{
		{
			Id:          "1-1",
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

	actions2 := []models.Action{
		{
			Id:          "2-1",
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

	items, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	// With significantly different configs, should not count as highly similar
	if len(items) != 0 {
		t.Errorf("Expected 0 for very different configurations, got %d", len(items))
	}
}

func TestFindActionsWithSimilarConfigurationsNoHashes(t *testing.T) {
	actions1 := []models.Action{
		{
			Id:          "1-1",
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
	}

	actions2 := []models.Action{
		{
			Id:          "2-1",
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
	}

	// Don't generate hashes - test nil hash handling

	items, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	if len(items) != 0 {
		t.Errorf("Expected 0 when hashes are not generated, got %d", len(items))
	}
}

func TestFindActionsWithSimilarConfigurationsMultiplePairs(t *testing.T) {
	actions1 := []models.Action{
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

	actions2 := []models.Action{
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

	items, _ := FindActionsWithSimilarConfigurations(actions1, actions2)

	// 2 actions in actions1 × 2 actions in actions2 = 4 comparisons
	// All should be highly similar, so count should be 8 (2 per similar pair)
	if len(items) < 4 {
		t.Errorf("Expected at least 4 (2x2x2) similar configurations, got %d", len(items))
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

func TestGenerateReportFromWorkflowsDissimilarSteps3(t *testing.T) {
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
      # It doesn't take much to make these dissimilar
      - name: Download JUnit Summary from Previous Workflow 2
        id: download-artifact-2
        uses: dawidd6/action-download-artifact@v6
        with:
          workflow_conclusion: success 
          name: junit-test-summary-2
          if_no_artifact_found: warn
          branch: main-2
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

func TestGenerateReportFromWorkflowsDifferentUsesSteps(t *testing.T) {
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
        uses: dawidd6/action-download-artifact-2@v6
        with:
          workflow_conclusion: success
          name: junit-test-summary
          if_no_artifact_found: warn
          branch: main
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
      - name: Script edited
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
          if_no_artifact_found: warn2
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
func TestGetActionAuthorsFromActionsList(t *testing.T) {
	tests := []struct {
		name        string
		actionsList [][]models.Action
		expected    []string
	}{
		{
			name:        "nil actionsList",
			actionsList: nil,
			expected:    []string{},
		},
		{
			name:        "empty actionsList",
			actionsList: [][]models.Action{},
			expected:    []string{},
		},
		{
			name: "single action from actions org",
			actionsList: [][]models.Action{
				{
					{Uses: "actions/checkout@v4"},
				},
			},
			expected: []string{"actions"},
		},
		{
			name: "multiple actions from same org",
			actionsList: [][]models.Action{
				{
					{Uses: "actions/checkout@v4"},
					{Uses: "actions/setup-node@v3"},
					{Uses: "actions/cache@v3"},
				},
			},
			expected: []string{"actions"},
		},
		{
			name: "actions from different orgs",
			actionsList: [][]models.Action{
				{
					{Uses: "actions/checkout@v4"},
					{Uses: "docker/build-push-action@v5"},
					{Uses: "hashicorp/setup-terraform@v2"},
				},
			},
			expected: []string{"actions", "docker", "hashicorp"},
		},
		{
			name: "built-in steps (empty Uses)",
			actionsList: [][]models.Action{
				{
					{Uses: ""},
					{Uses: "actions/checkout@v4"},
				},
			},
			expected: []string{BuiltInStep, "actions"},
		},
		{
			name: "multiple built-in steps",
			actionsList: [][]models.Action{
				{
					{Uses: ""},
					{Uses: ""},
					{Uses: "actions/checkout@v4"},
				},
			},
			expected: []string{BuiltInStep, "actions"},
		},
		{
			name: "multiple workflows with same authors",
			actionsList: [][]models.Action{
				{
					{Uses: "actions/checkout@v4"},
					{Uses: "docker/build-push-action@v5"},
				},
				{
					{Uses: "actions/setup-node@v3"},
					{Uses: "docker/login-action@v2"},
				},
			},
			expected: []string{"actions", "docker"},
		},
		{
			name: "multiple workflows with different authors",
			actionsList: [][]models.Action{
				{
					{Uses: "actions/checkout@v4"},
				},
				{
					{Uses: "docker/build-push-action@v5"},
				},
				{
					{Uses: "hashicorp/setup-terraform@v2"},
				},
			},
			expected: []string{"actions", "docker", "hashicorp"},
		},
		{
			name: "actions with three-part names",
			actionsList: [][]models.Action{
				{
					{Uses: "octocat/hello-world-docker-action@v1"},
					{Uses: "github/codeql-action/init@v2"},
				},
			},
			expected: []string{"octocat", "github"},
		},
		{
			name: "duplicate authors across workflows",
			actionsList: [][]models.Action{
				{
					{Uses: "actions/checkout@v4"},
					{Uses: "actions/setup-node@v3"},
				},
				{
					{Uses: "actions/cache@v3"},
					{Uses: "actions/upload-artifact@v3"},
				},
			},
			expected: []string{"actions"},
		},
		{
			name: "local actions (relative paths)",
			actionsList: [][]models.Action{
				{
					{Uses: "./.github/actions/my-action"},
					{Uses: "./local-action"},
					{Uses: "actions/checkout@v4"},
				},
			},
			expected: []string{".", "actions"},
		},
		{
			name: "actions with no slashes",
			actionsList: [][]models.Action{
				{
					{Uses: "standalone-action"},
					{Uses: "actions/checkout@v4"},
				},
			},
			expected: []string{"standalone-action", "actions"},
		},
		{
			name: "mixed built-in and regular actions",
			actionsList: [][]models.Action{
				{
					{Uses: ""},
					{Uses: "actions/checkout@v4"},
					{Uses: ""},
					{Uses: "docker/build-push-action@v5"},
				},
			},
			expected: []string{BuiltInStep, "actions", "docker"},
		},
		{
			name: "empty workflows in list",
			actionsList: [][]models.Action{
				{},
				{
					{Uses: "actions/checkout@v4"},
				},
				{},
			},
			expected: []string{"actions"},
		},
		{
			name: "complex scenario with all types",
			actionsList: [][]models.Action{
				{
					{Uses: ""},
					{Uses: "actions/checkout@v4"},
					{Uses: "docker/build-push-action@v5"},
				},
				{
					{Uses: "hashicorp/setup-terraform@v2"},
					{Uses: "actions/setup-go@v4"},
				},
				{
					{Uses: "./.github/actions/custom"},
					{Uses: ""},
				},
			},
			expected: []string{BuiltInStep, "actions", "docker", "hashicorp", "."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetActionAuthorsFromActionsList(tt.actionsList)
			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("GetActionAuthorsFromActionsList() returned %d authors, expected %d", len(result), len(tt.expected))
				t.Logf("Result: %v", result)
				t.Logf("Expected: %v", tt.expected)
				return
			}
			// Check that all expected authors are present (order doesn't matter for uniqueness)
			for _, expectedAuthor := range tt.expected {
				found := false
				for _, resultAuthor := range result {
					if resultAuthor == expectedAuthor {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetActionAuthorsFromActionsList() missing expected author %q", expectedAuthor)
					t.Logf("Result: %v", result)
					t.Logf("Expected: %v", tt.expected)
				}
			}
			// Check that no unexpected authors are present
			for _, resultAuthor := range result {
				found := false
				for _, expectedAuthor := range tt.expected {
					if resultAuthor == expectedAuthor {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetActionAuthorsFromActionsList() has unexpected author %q", resultAuthor)
					t.Logf("Result: %v", result)
					t.Logf("Expected: %v", tt.expected)
				}
			}
		})
	}
}
func TestGetActionAuthorsFromActionsListUniqueness(t *testing.T) {
	// Test that the function returns unique authors only
	actionsList := [][]models.Action{
		{
			{Uses: "actions/checkout@v4"},
			{Uses: "actions/setup-node@v3"},
			{Uses: "actions/cache@v3"},
		},
		{
			{Uses: "actions/checkout@v3"}, // Same author, different version
			{Uses: "actions/upload-artifact@v3"},
		},
	}
	result := GetActionAuthorsFromActionsList(actionsList)
	// Should only have "actions" once
	if len(result) != 1 {
		t.Errorf("GetActionAuthorsFromActionsList() returned %d authors, expected 1", len(result))
		t.Logf("Result: %v", result)
	}
	if result[0] != "actions" {
		t.Errorf("GetActionAuthorsFromActionsList() returned %q, expected 'actions'", result[0])
	}
}
func TestGetActionAuthorsFromActionsListEmptyStrings(t *testing.T) {
	// Test handling of empty Uses strings
	actionsList := [][]models.Action{
		{
			{Uses: ""},
			{Uses: ""},
			{Uses: ""},
		},
	}
	result := GetActionAuthorsFromActionsList(actionsList)
	// Should only have BuiltInStep once
	if len(result) != 1 {
		t.Errorf("GetActionAuthorsFromActionsList() returned %d authors, expected 1", len(result))
		t.Logf("Result: %v", result)
	}
	if result[0] != BuiltInStep {
		t.Errorf("GetActionAuthorsFromActionsList() returned %q, expected %q", result[0], BuiltInStep)
	}
}
