package githubapi

import (
	"net/http"
	"testing"

	"github.com/google/go-github/v57/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestFindWorkflows_Success(t *testing.T) {
	// Arrange
	yamlFile := github.RepositoryContent{
		Name: github.String("build.yml"),
		Type: github.String("file"),
	}
	yamlFile2 := github.RepositoryContent{
		Name: github.String("test.yaml"),
		Type: github.String("file"),
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{yamlFile, yamlFile2},
			nil,
		),
	)

	client := github.NewClient(mockedHTTPClient)

	// Act
	workflows := FindWorkflows(client, "owner/repo")

	// Assert
	if len(workflows) != 2 {
		t.Errorf("Expected 2 workflows, got %d", len(workflows))
	}

	if workflows[0] != "build.yml" {
		t.Errorf("Expected first workflow to be 'build.yml', got '%s'", workflows[0])
	}

	if workflows[1] != "test.yaml" {
		t.Errorf("Expected second workflow to be 'test.yaml', got '%s'", workflows[1])
	}
}

func TestFindWorkflows_MixedFiles(t *testing.T) {
	// Arrange - mix of YAML files and other files
	yamlFile := github.RepositoryContent{
		Name: github.String("build.yml"),
		Type: github.String("file"),
	}
	yamlFile2 := github.RepositoryContent{
		Name: github.String("deploy.YAML"),
		Type: github.String("file"),
	}
	txtFile := github.RepositoryContent{
		Name: github.String("readme.txt"),
		Type: github.String("file"),
	}
	dirEntry := github.RepositoryContent{
		Name: github.String("subdir"),
		Type: github.String("dir"),
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{yamlFile, yamlFile2, txtFile, dirEntry},
			nil,
		),
	)

	client := github.NewClient(mockedHTTPClient)

	// Act
	workflows := FindWorkflows(client, "owner/repo")

	// Assert
	if len(workflows) != 2 {
		t.Errorf("Expected 2 workflows (only YAML files), got %d", len(workflows))
	}

	expectedFiles := map[string]bool{
		"build.yml":   false,
		"deploy.YAML": false,
	}

	for _, wf := range workflows {
		if _, exists := expectedFiles[wf]; exists {
			expectedFiles[wf] = true
		} else {
			t.Errorf("Unexpected workflow file: %s", wf)
		}
	}

	for file, found := range expectedFiles {
		if !found {
			t.Errorf("Expected workflow file '%s' not found", file)
		}
	}
}

func TestFindWorkflows_EmptyDirectory(t *testing.T) {
	// Arrange
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{},
			nil,
		),
	)

	client := github.NewClient(mockedHTTPClient)

	// Act
	workflows := FindWorkflows(client, "owner/repo")

	// Assert
	if len(workflows) != 0 {
		t.Errorf("Expected 0 workflows for empty directory, got %d", len(workflows))
	}
}

func TestFindWorkflows_InvalidRepoFormat(t *testing.T) {
	// Arrange
	client := github.NewClient(nil)

	// Act
	workflows := FindWorkflows(client, "invalid-repo-format")

	// Assert
	if len(workflows) != 0 {
		t.Errorf("Expected 0 workflows for invalid repo format, got %d", len(workflows))
	}
}

func TestFindWorkflows_APIError(t *testing.T) {
	// Arrange - mock an API error
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
	workflows := FindWorkflows(client, "owner/repo")

	// Assert
	if len(workflows) != 0 {
		t.Errorf("Expected 0 workflows when API returns error, got %d", len(workflows))
	}
}

func TestFindWorkflows_OnlyDirectories(t *testing.T) {
	// Arrange - only directories, no files
	dirEntry1 := github.RepositoryContent{
		Name: github.String("subdir1"),
		Type: github.String("dir"),
	}
	dirEntry2 := github.RepositoryContent{
		Name: github.String("subdir2"),
		Type: github.String("dir"),
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{dirEntry1, dirEntry2},
			nil,
		),
	)

	client := github.NewClient(mockedHTTPClient)

	// Act
	workflows := FindWorkflows(client, "owner/repo")

	// Assert
	if len(workflows) != 0 {
		t.Errorf("Expected 0 workflows when directory contains only subdirectories, got %d", len(workflows))
	}
}

func TestFindWorkflows_CaseInsensitiveExtension(t *testing.T) {
	// Arrange - various case combinations
	file1 := github.RepositoryContent{
		Name: github.String("test.yml"),
		Type: github.String("file"),
	}
	file2 := github.RepositoryContent{
		Name: github.String("test.YML"),
		Type: github.String("file"),
	}
	file3 := github.RepositoryContent{
		Name: github.String("test.yaml"),
		Type: github.String("file"),
	}
	file4 := github.RepositoryContent{
		Name: github.String("test.YAML"),
		Type: github.String("file"),
	}
	file5 := github.RepositoryContent{
		Name: github.String("test.Yml"),
		Type: github.String("file"),
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{file1, file2, file3, file4, file5},
			nil,
		),
	)

	client := github.NewClient(mockedHTTPClient)

	// Act
	workflows := FindWorkflows(client, "owner/repo")

	// Assert
	if len(workflows) != 5 {
		t.Errorf("Expected 5 workflows (all case variations), got %d", len(workflows))
	}
}

func TestFindWorkflows_RepoWithOrganization(t *testing.T) {
	// Arrange
	yamlFile := github.RepositoryContent{
		Name: github.String("ci.yml"),
		Type: github.String("file"),
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{yamlFile},
			nil,
		),
	)

	client := github.NewClient(mockedHTTPClient)

	// Act
	workflows := FindWorkflows(client, "myorg/myrepo")

	// Assert
	if len(workflows) != 1 {
		t.Errorf("Expected 1 workflow, got %d", len(workflows))
	}

	if len(workflows) > 0 && workflows[0] != "ci.yml" {
		t.Errorf("Expected workflow to be 'ci.yml', got '%s'", workflows[0])
	}
}

func TestFindWorkflows_NoWorkflowsDirectory(t *testing.T) {
	// Arrange - simulate .github/workflows directory doesn't exist
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
	workflows := FindWorkflows(client, "owner/repo-without-workflows")

	// Assert
	if len(workflows) != 0 {
		t.Errorf("Expected 0 workflows when .github/workflows doesn't exist, got %d", len(workflows))
	}
}

func TestFindWorkflows_ContextUsage(t *testing.T) {
	// This test verifies that the function uses context.Background()
	// The actual implementation uses context.Background(), which is fine for this function
	// This test just ensures the function can be called and works with the context

	yamlFile := github.RepositoryContent{
		Name: github.String("test.yml"),
		Type: github.String("file"),
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposContentsByOwnerByRepoByPath,
			[]github.RepositoryContent{yamlFile},
			nil,
		),
	)

	client := github.NewClient(mockedHTTPClient)

	// Act
	workflows := FindWorkflows(client, "owner/repo")

	// Assert
	if len(workflows) != 1 {
		t.Errorf("Expected 1 workflow, got %d", len(workflows))
	}
}
