package githubapi

import (
	"net/http"
	"testing"

	"github.com/google/go-github/v57/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestFindContributorsToWorkflow_Success(t *testing.T) {
	// Arrange
	author1 := "John Doe"
	author2 := "Jane Smith"
	commit1 := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{
				Name: &author1,
			},
		},
	}
	commit2 := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{
				Name: &author2,
			},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			[]github.RepositoryCommit{commit1, commit2},
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "ci.yml")
	// Assert
	if len(contributors) != 2 {
		t.Errorf("Expected 2 contributors, got %d", len(contributors))
	}
	expectedContributors := map[string]bool{
		"John Doe":   false,
		"Jane Smith": false,
	}
	for _, contributor := range contributors {
		if _, exists := expectedContributors[contributor]; exists {
			expectedContributors[contributor] = true
		} else {
			t.Errorf("Unexpected contributor: %s", contributor)
		}
	}
	for name, found := range expectedContributors {
		if !found {
			t.Errorf("Expected contributor '%s' not found", name)
		}
	}
}
func TestFindContributorsToWorkflow_SingleContributor(t *testing.T) {
	// Arrange
	author := "Alice Developer"
	commit := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{
				Name: &author,
			},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			[]github.RepositoryCommit{commit},
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "deploy.yaml")
	// Assert
	if len(contributors) != 1 {
		t.Errorf("Expected 1 contributor, got %d", len(contributors))
	}
	if len(contributors) > 0 && contributors[0] != "Alice Developer" {
		t.Errorf("Expected contributor 'Alice Developer', got '%s'", contributors[0])
	}
}
func TestFindContributorsToWorkflow_DuplicateContributors(t *testing.T) {
	// Arrange - same contributor made multiple commits
	author := "Bob Builder"
	commit1 := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{
				Name: &author,
			},
		},
	}
	commit2 := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{
				Name: &author,
			},
		},
	}
	commit3 := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{
				Name: &author,
			},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			[]github.RepositoryCommit{commit1, commit2, commit3},
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "test.yml")
	// Assert - should deduplicate to single contributor
	if len(contributors) != 1 {
		t.Errorf("Expected 1 unique contributor, got %d", len(contributors))
	}
	if len(contributors) > 0 && contributors[0] != "Bob Builder" {
		t.Errorf("Expected contributor 'Bob Builder', got '%s'", contributors[0])
	}
}
func TestFindContributorsToWorkflow_MultipleContributorsWithDuplicates(t *testing.T) {
	// Arrange - mix of contributors with some duplicates
	author1 := "Charlie Code"
	author2 := "Diana Dev"
	commits := []github.RepositoryCommit{
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author1},
			},
		},
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author2},
			},
		},
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author1},
			},
		},
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author2},
			},
		},
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author1},
			},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			commits,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "workflow.yml")
	// Assert - should have 2 unique contributors
	if len(contributors) != 2 {
		t.Errorf("Expected 2 unique contributors, got %d", len(contributors))
	}
	expectedContributors := map[string]bool{
		"Charlie Code": false,
		"Diana Dev":    false,
	}
	for _, contributor := range contributors {
		if _, exists := expectedContributors[contributor]; exists {
			expectedContributors[contributor] = true
		}
	}
	for name, found := range expectedContributors {
		if !found {
			t.Errorf("Expected contributor '%s' not found", name)
		}
	}
}
func TestFindContributorsToWorkflow_NoCommits(t *testing.T) {
	// Arrange - empty commit list
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			[]github.RepositoryCommit{},
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "new-workflow.yml")
	// Assert
	if len(contributors) != 0 {
		t.Errorf("Expected 0 contributors for workflow with no commits, got %d", len(contributors))
	}
}
func TestFindContributorsToWorkflow_InvalidRepoFormat(t *testing.T) {
	// Arrange
	client := github.NewClient(nil)
	// Act
	contributors := FindContributorsToWorkflow(client, "invalid-repo-format", "ci.yml")
	// Assert
	if len(contributors) != 0 {
		t.Errorf("Expected 0 contributors for invalid repo format, got %d", len(contributors))
	}
}
func TestFindContributorsToWorkflow_RepoWithoutSlash(t *testing.T) {
	// Arrange
	client := github.NewClient(nil)
	// Act
	contributors := FindContributorsToWorkflow(client, "justreponame", "ci.yml")
	// Assert
	if len(contributors) != 0 {
		t.Errorf("Expected 0 contributors for repo without slash, got %d", len(contributors))
	}
}
func TestFindContributorsToWorkflow_APIError(t *testing.T) {
	// Arrange - mock API returning error
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposCommitsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "ci.yml")
	// Assert
	if len(contributors) != 0 {
		t.Errorf("Expected 0 contributors when API returns error, got %d", len(contributors))
	}
}
func TestFindContributorsToWorkflow_FileNotFound(t *testing.T) {
	// Arrange - mock API returning 404
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposCommitsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "nonexistent.yml")
	// Assert
	if len(contributors) != 0 {
		t.Errorf("Expected 0 contributors when workflow file not found, got %d", len(contributors))
	}
}
func TestFindContributorsToWorkflow_Unauthorized(t *testing.T) {
	// Arrange - mock API returning 403
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposCommitsByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			}),
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/private-repo", "ci.yml")
	// Assert
	if len(contributors) != 0 {
		t.Errorf("Expected 0 contributors when unauthorized, got %d", len(contributors))
	}
}
func TestFindContributorsToWorkflow_CommitsWithNilAuthor(t *testing.T) {
	// Arrange - some commits have nil author (should be filtered out)
	author := "Valid Author"
	commits := []github.RepositoryCommit{
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author},
			},
		},
		{
			Commit: &github.Commit{
				Author: nil, // Nil author
			},
		},
		{
			Commit: nil, // Nil commit
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			commits,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "ci.yml")
	// Assert - should only include the valid author
	if len(contributors) != 1 {
		t.Errorf("Expected 1 contributor (filtering out nil authors), got %d", len(contributors))
	}
	if len(contributors) > 0 && contributors[0] != "Valid Author" {
		t.Errorf("Expected contributor 'Valid Author', got '%s'", contributors[0])
	}
}
func TestFindContributorsToWorkflow_CommitsWithNilName(t *testing.T) {
	// Arrange - some commits have author with nil name
	author := "Real Name"
	commits := []github.RepositoryCommit{
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author},
			},
		},
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: nil}, // Nil name
			},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			commits,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "ci.yml")
	// Assert - should only include the valid author
	if len(contributors) != 1 {
		t.Errorf("Expected 1 contributor (filtering out nil names), got %d", len(contributors))
	}
	if len(contributors) > 0 && contributors[0] != "Real Name" {
		t.Errorf("Expected contributor 'Real Name', got '%s'", contributors[0])
	}
}
func TestFindContributorsToWorkflow_ManyContributors(t *testing.T) {
	// Arrange - workflow with many contributors
	var commits []github.RepositoryCommit
	expectedNames := []string{
		"Alice", "Bob", "Charlie", "Diana", "Eve",
		"Frank", "Grace", "Henry", "Ivy", "Jack",
	}
	for _, name := range expectedNames {
		nameCopy := name
		commits = append(commits, github.RepositoryCommit{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &nameCopy},
			},
		})
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			commits,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "ci.yml")
	// Assert
	if len(contributors) != 10 {
		t.Errorf("Expected 10 contributors, got %d", len(contributors))
	}
	contributorMap := make(map[string]bool)
	for _, contributor := range contributors {
		contributorMap[contributor] = true
	}
	for _, expectedName := range expectedNames {
		if !contributorMap[expectedName] {
			t.Errorf("Expected contributor '%s' not found", expectedName)
		}
	}
}
func TestFindContributorsToWorkflow_YamlExtension(t *testing.T) {
	// Arrange - test with .yaml extension
	author := "YAML User"
	commit := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{Name: &author},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			[]github.RepositoryCommit{commit},
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "deploy.yaml")
	// Assert
	if len(contributors) != 1 {
		t.Errorf("Expected 1 contributor, got %d", len(contributors))
	}
	if len(contributors) > 0 && contributors[0] != "YAML User" {
		t.Errorf("Expected contributor 'YAML User', got '%s'", contributors[0])
	}
}
func TestFindContributorsToWorkflow_HTTPSRepoURL(t *testing.T) {
	// Arrange - test with HTTPS repo URL format
	author := "URL Format User"
	commit := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{Name: &author},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			[]github.RepositoryCommit{commit},
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "https://github.com/owner/repo", "ci.yml")
	// Assert
	if len(contributors) != 1 {
		t.Errorf("Expected 1 contributor, got %d", len(contributors))
	}
	if len(contributors) > 0 && contributors[0] != "URL Format User" {
		t.Errorf("Expected contributor 'URL Format User', got '%s'", contributors[0])
	}
}
func TestFindContributorsToWorkflow_ContributorWithSpecialCharacters(t *testing.T) {
	// Arrange - test with special characters in names
	author1 := "José García"
	author2 := "François Müller"
	author3 := "李明"
	commits := []github.RepositoryCommit{
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author1},
			},
		},
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author2},
			},
		},
		{
			Commit: &github.Commit{
				Author: &github.CommitAuthor{Name: &author3},
			},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			commits,
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "owner/repo", "ci.yml")
	// Assert
	if len(contributors) != 3 {
		t.Errorf("Expected 3 contributors, got %d", len(contributors))
	}
	expectedContributors := map[string]bool{
		"José García":     false,
		"François Müller": false,
		"李明":              false,
	}
	for _, contributor := range contributors {
		if _, exists := expectedContributors[contributor]; exists {
			expectedContributors[contributor] = true
		}
	}
	for name, found := range expectedContributors {
		if !found {
			t.Errorf("Expected contributor '%s' not found", name)
		}
	}
}
func TestFindContributorsToWorkflow_OrganizationRepo(t *testing.T) {
	// Arrange - test with organization repository
	author := "Org Developer"
	commit := github.RepositoryCommit{
		Commit: &github.Commit{
			Author: &github.CommitAuthor{Name: &author},
		},
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposCommitsByOwnerByRepo,
			[]github.RepositoryCommit{commit},
			nil,
		),
	)
	client := github.NewClient(mockedHTTPClient)
	// Act
	contributors := FindContributorsToWorkflow(client, "my-organization/my-project", "ci.yml")
	// Assert
	if len(contributors) != 1 {
		t.Errorf("Expected 1 contributor, got %d", len(contributors))
	}
	if len(contributors) > 0 && contributors[0] != "Org Developer" {
		t.Errorf("Expected contributor 'Org Developer', got '%s'", contributors[0])
	}
}
