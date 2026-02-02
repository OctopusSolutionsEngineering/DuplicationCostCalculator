package workflows

import (
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

	actions := ParseWorkflow(workflowYAML)

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
	actions := ParseWorkflow(invalidYAML)

	if len(actions) != 0 {
		t.Errorf("Expected 0 actions for invalid YAML, got %d", len(actions))
	}
}

func TestParseWorkflowNoJobs(t *testing.T) {
	workflowYAML := `
name: Empty Workflow
on: [push]
`
	actions := ParseWorkflow(workflowYAML)

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
	actions := ParseWorkflow(workflowYAML)

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

	count := FindActionsWithDifferentVersions(actions1, actions2)

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

	count := FindActionsWithDifferentVersions(actions1, actions2)

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

	count := FindActionsWithDifferentVersions(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 actions with different versions when all match, got %d", count)
	}
}

func TestFindActionsWithDifferentVersionsEmptyLists(t *testing.T) {
	actions1 := []Action{}
	actions2 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v4"},
	}

	count := FindActionsWithDifferentVersions(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 actions with different versions when one list is empty, got %d", count)
	}
}

func TestFindActionsWithDifferentVersionsMultipleSameAction(t *testing.T) {
	actions1 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v3"},
		{Uses: "actions/checkout", UsesVersion: "v3"},
	}

	actions2 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v4"},
		{Uses: "actions/checkout", UsesVersion: "v4"},
	}

	count := FindActionsWithDifferentVersions(actions1, actions2)

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

	count := FindActionsWithSimilarConfigurations(actions1, actions2)

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

	count := FindActionsWithSimilarConfigurations(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 similar configurations when action names differ, got %d", count)
	}
}

func TestFindActionsWithSimilarConfigurationsEmptyLists(t *testing.T) {
	actions1 := []Action{}
	actions2 := []Action{
		{Uses: "actions/checkout", UsesVersion: "v4"},
	}

	count := FindActionsWithSimilarConfigurations(actions1, actions2)

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

	count := FindActionsWithSimilarConfigurations(actions1, actions2)

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

	count := FindActionsWithSimilarConfigurations(actions1, actions2)

	if count != 0 {
		t.Errorf("Expected 0 when hashes are not generated, got %d", count)
	}
}

func TestFindActionsWithSimilarConfigurationsMultiplePairs(t *testing.T) {
	actions1 := []Action{
		{
			Uses:        "actions/checkout",
			UsesVersion: "v4",
			With: map[string]string{
				"fetch-depth": "0",
			},
		},
		{
			Uses:        "actions/checkout",
			UsesVersion: "v3",
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
		{
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

	count := FindActionsWithSimilarConfigurations(actions1, actions2)

	// 2 actions in actions1 × 2 actions in actions2 = 4 comparisons
	// All should be highly similar, so count should be 8 (2 per similar pair)
	if count < 4 {
		t.Errorf("Expected at least 4 (2x2x2) similar configurations, got %d", count)
	}
}
