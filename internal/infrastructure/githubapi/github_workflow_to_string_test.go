package githubapi

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/google/go-github/v57/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestWorkflowToString_Success(t *testing.T) {
	// Arrange
	workflowContent := `name: CI
on: [push, pull_request]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        run: go test ./...`
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("ci.yml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "ci.yml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected workflow content to match.\nExpected:\n%s\n\nGot:\n%s", workflowContent, result)
	}
}
func TestWorkflowToString_YamlExtension(t *testing.T) {
	// Arrange
	workflowContent := `name: Deploy
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2`
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("deploy.yaml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "deploy.yaml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected workflow content to match.\nExpected:\n%s\n\nGot:\n%s", workflowContent, result)
	}
}
func TestWorkflowToString_ComplexWorkflow(t *testing.T) {
	// Arrange - test with a more complex workflow containing special characters
	workflowContent := `name: Complex Workflow
on:
  schedule:
    - cron: '0 0 * * *'
  workflow_dispatch:
    inputs:
      environment:
        description: 'Environment to deploy to'
        required: true
        default: 'staging'
env:
  NODE_VERSION: '18.x'
  PYTHON_VERSION: '3.11'
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node: [16, 18, 20]
    steps:
      - name: Checkout
        uses: actions/checkout@v3
      - name: Setup Node ${{ matrix.node }}
        uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node }}
      - name: Install & Test
        run: |
          npm ci
          npm test
          echo "Tests passed!"
  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy
        run: echo "Deploying..."`
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("complex.yml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "complex.yml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected workflow content to match.\nExpected:\n%s\n\nGot:\n%s", workflowContent, result)
	}
}
func TestWorkflowToString_EmptyWorkflow(t *testing.T) {
	// Arrange - empty workflow file
	workflowContent := ""
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("empty.yml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "empty.yml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}
func TestWorkflowToString_InvalidRepoFormat(t *testing.T) {
	// Arrange
	client := github.NewClient(nil)
	// Act
	result := WorkflowToString(client, "invalid-repo-format", "ci.yml")
	// Assert
	if result != "" {
		t.Errorf("Expected empty string for invalid repo format, got '%s'", result)
	}
}
func TestWorkflowToString_RepoWithoutSlash(t *testing.T) {
	// Arrange
	client := github.NewClient(nil)
	// Act
	result := WorkflowToString(client, "justreponame", "ci.yml")
	// Assert
	if result != "" {
		t.Errorf("Expected empty string for repo without slash, got '%s'", result)
	}
}
func TestWorkflowToString_FileNotFound(t *testing.T) {
	// Arrange - mock API returning 404
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposContentsByOwnerByRepoByPath,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "nonexistent.yml")
	// Assert
	if result != "" {
		t.Errorf("Expected empty string when file not found, got '%s'", result)
	}
}
func TestWorkflowToString_APIError(t *testing.T) {
	// Arrange - mock API returning 500 internal server error
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposContentsByOwnerByRepoByPath,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "ci.yml")
	// Assert
	if result != "" {
		t.Errorf("Expected empty string when API returns error, got '%s'", result)
	}
}
func TestWorkflowToString_Unauthorized(t *testing.T) {
	// Arrange - mock API returning 403 forbidden
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposContentsByOwnerByRepoByPath,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			}),
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/private-repo", "ci.yml")
	// Assert
	if result != "" {
		t.Errorf("Expected empty string when unauthorized, got '%s'", result)
	}
}
func TestWorkflowToString_WorkflowWithNestedPath(t *testing.T) {
	// Arrange - ensure the function correctly constructs .github/workflows/ path
	workflowContent := `name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Testing"`
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("test.yml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "myorg/myrepo", "test.yml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected workflow content to match.\nExpected:\n%s\n\nGot:\n%s", workflowContent, result)
	}
}
func TestWorkflowToString_MultilineWorkflow(t *testing.T) {
	// Arrange - test with multiline strings and various YAML features
	workflowContent := `name: Multiline Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Multiline command
        run: |
          echo "Line 1"
          echo "Line 2"
          echo "Line 3"
      - name: Another step
        run: >
          This is a folded
          multiline string`
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("multiline.yml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "multiline.yml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected workflow content to match.\nExpected:\n%s\n\nGot:\n%s", workflowContent, result)
	}
}
func TestWorkflowToString_UnicodeContent(t *testing.T) {
	// Arrange - test with Unicode characters
	workflowContent := `name: Unicode Test 🚀
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Test with émojis 🎉
        run: echo "Testing with Unicode: 你好世界 🌍"`
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("unicode.yml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "unicode.yml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected workflow content to match.\nExpected:\n%s\n\nGot:\n%s", workflowContent, result)
	}
}
func TestWorkflowToString_RepoWithHTTPSURL(t *testing.T) {
	// Arrange - test with repo in HTTPS URL format
	workflowContent := `name: URL Format Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "test"`
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("test.yml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "https://github.com/owner/repo", "test.yml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected workflow content to match.\nExpected:\n%s\n\nGot:\n%s", workflowContent, result)
	}
}
func TestWorkflowToString_WorkflowWithSpecialCharacters(t *testing.T) {
	// Arrange - test with special characters in workflow content
	workflowContent := `name: Special Characters
on: [push]
env:
  SPECIAL: 'value with "quotes" and $variables'
  PATH_VAR: '/path/to/file'
  REGEX: '^[a-z]+$'
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Test special chars
        run: echo "Testing: !@#$%^&*()_+-=[]{}|;:',.<>?/"`
	encodedContent := base64.StdEncoding.EncodeToString([]byte(workflowContent))
	fileContent := github.RepositoryContent{
		Name:     github.String("special.yml"),
		Type:     github.String("file"),
		Content:  &encodedContent,
		Encoding: github.String("base64"),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			fileContent,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	result := WorkflowToString(client, "owner/repo", "special.yml")
	// Assert
	if result != workflowContent {
		t.Errorf("Expected workflow content to match.\nExpected:\n%s\n\nGot:\n%s", workflowContent, result)
	}
}
