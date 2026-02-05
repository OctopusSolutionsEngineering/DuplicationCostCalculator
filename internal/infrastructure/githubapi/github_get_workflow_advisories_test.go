package githubapi

import (
	"testing"

	"github.com/google/go-github/v57/github"
)

func TestGetWorkflowAdvisories_InvalidRepoFormat(t *testing.T) {
	// Arrange
	client := github.NewClient(nil)

	// Act
	advisories := GetWorkflowAdvisories(client, "invalid-repo-format")

	// Assert
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories for invalid repo format, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_RepoWithoutSlash(t *testing.T) {
	// Arrange
	client := github.NewClient(nil)

	// Act
	advisories := GetWorkflowAdvisories(client, "justreponame")

	// Assert
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories for repo without slash, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_EmptyRepo(t *testing.T) {
	// Arrange
	client := github.NewClient(nil)

	// Act
	advisories := GetWorkflowAdvisories(client, "")

	// Assert
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories for empty repo string, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_HTTPSRepoURL(t *testing.T) {
	// Arrange - test with HTTPS repo URL format (this will fail the API call but should pass SplitRepo)
	client := github.NewClient(nil)

	// Act
	advisories := GetWorkflowAdvisories(client, "https://github.com/owner/repo")

	// Assert - Should return empty due to nil client causing API error
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories with nil client, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_RepoWithDashes(t *testing.T) {
	// Arrange - test repo names with dashes (valid format but will fail at API call)
	client := github.NewClient(nil)

	// Act
	advisories := GetWorkflowAdvisories(client, "my-org-name/my-repo-name")

	// Assert - Should return empty due to nil client causing API error
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories with nil client, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_RepoWithSpecialChars(t *testing.T) {
	// Arrange - test with special characters in repo name
	client := github.NewClient(nil)

	// Act
	advisories := GetWorkflowAdvisories(client, "owner/repo@#$%")

	// Assert - Should return empty due to nil client
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories with nil client, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_NilClient(t *testing.T) {
	// Arrange - test with nil client
	var client *github.Client = nil

	// Act
	advisories := GetWorkflowAdvisories(client, "owner/repo")

	// Assert - Should return empty due to nil client
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories with nil client, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_MultipleSlashes(t *testing.T) {
	// Arrange - test with multiple slashes in repo string
	client := github.NewClient(nil)

	// Act
	advisories := GetWorkflowAdvisories(client, "owner/repo/extra/path")

	// Assert - SplitRepo should handle this, returning empty on API error
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_RepoWithWhitespace(t *testing.T) {
	// Arrange - test with whitespace in repo name
	client := github.NewClient(nil)

	// Act
	advisories := GetWorkflowAdvisories(client, "  owner / repo  ")

	// Assert - Should handle whitespace gracefully
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories with whitespace, got %d", len(advisories))
	}
}

func TestGetWorkflowAdvisories_VeryLongRepoName(t *testing.T) {
	// Arrange - test with very long repo name
	client := github.NewClient(nil)
	longName := "owner/" + string(make([]byte, 500))

	// Act
	advisories := GetWorkflowAdvisories(client, longName)

	// Assert - Should handle long names without panic
	if len(advisories) != 0 {
		t.Errorf("Expected 0 advisories with long name, got %d", len(advisories))
	}
}
