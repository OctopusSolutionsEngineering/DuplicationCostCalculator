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
          
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
`

	actions := ParseWorkflow(workflowYAML)

	if len(actions) != 4 {
		t.Errorf("Expected 4 actions, got %d", len(actions))
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

	// Test fourth action (no version specified)
	if actions[3].Uses != "actions/checkout" {
		t.Errorf("Expected fourth action to be 'actions/checkout', got '%s'", actions[3].Uses)
	}
	if actions[3].UsesVersion != "v3" {
		t.Errorf("Expected fourth action version to be 'v3', got '%s'", actions[3].UsesVersion)
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
