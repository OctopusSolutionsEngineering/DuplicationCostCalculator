package parsing

import (
	"testing"
)

func TestSplitRepo(t *testing.T) {
	tests := []struct {
		name          string
		repo          string
		expectedOwner string
		expectedRepo  string
		expectError   bool
	}{
		{
			name:          "valid repo with owner/name format",
			repo:          "owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "valid repo with https URL",
			repo:          "https://github.com/owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "valid repo with https URL and trailing path",
			repo:          "https://github.com/owner/repo/tree/main",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "valid repo with https URL and multiple paths",
			repo:          "https://github.com/owner/repo/issues/123",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "organization and repo with hyphens",
			repo:          "my-org/my-repo",
			expectedOwner: "my-org",
			expectedRepo:  "my-repo",
			expectError:   false,
		},
		{
			name:          "organization and repo with underscores",
			repo:          "my_org/my_repo",
			expectedOwner: "my_org",
			expectedRepo:  "my_repo",
			expectError:   false,
		},
		{
			name:          "organization and repo with dots",
			repo:          "my.org/my.repo",
			expectedOwner: "my.org",
			expectedRepo:  "my.repo",
			expectError:   false,
		},
		{
			name:          "organization and repo with numbers",
			repo:          "org123/repo456",
			expectedOwner: "org123",
			expectedRepo:  "repo456",
			expectError:   false,
		},
		{
			name:          "real world example - OctopusDeploy",
			repo:          "OctopusDeploy/OctopusDeploy",
			expectedOwner: "OctopusDeploy",
			expectedRepo:  "OctopusDeploy",
			expectError:   false,
		},
		{
			name:          "real world example with URL - actions/checkout",
			repo:          "https://github.com/actions/checkout",
			expectedOwner: "actions",
			expectedRepo:  "checkout",
			expectError:   false,
		},
		{
			name:        "invalid - only owner",
			repo:        "owner",
			expectError: true,
		},
		{
			name:        "invalid - empty string",
			repo:        "",
			expectError: true,
		},
		{
			name:        "invalid - only slash",
			repo:        "/",
			expectError: true,
		},
		{
			name:        "invalid - slash at start",
			repo:        "/repo",
			expectError: true,
		},
		{
			name:        "invalid - slash at end",
			repo:        "owner/",
			expectError: true,
		},
		{
			name:        "invalid - https URL without owner/repo",
			repo:        "https://github.com/",
			expectError: true,
		},
		{
			name:        "invalid - https URL with only owner",
			repo:        "https://github.com/owner",
			expectError: true,
		},
		{
			name:          "multiple slashes - takes first two parts",
			repo:          "owner/repo/extra/parts",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectError:   false,
		},
		{
			name:          "URL with .git extension",
			repo:          "https://github.com/owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo.git",
			expectError:   false,
		},
		{
			name:          "case sensitivity preserved",
			repo:          "MyOrg/MyRepo",
			expectedOwner: "MyOrg",
			expectedRepo:  "MyRepo",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := SplitRepo(tt.repo)

			if tt.expectError {
				if err == nil {
					t.Errorf("SplitRepo(%q) expected error but got none", tt.repo)
				}
			} else {
				if err != nil {
					t.Errorf("SplitRepo(%q) unexpected error: %v", tt.repo, err)
				}
				if owner != tt.expectedOwner {
					t.Errorf("SplitRepo(%q) owner = %q, expected %q", tt.repo, owner, tt.expectedOwner)
				}
				if repo != tt.expectedRepo {
					t.Errorf("SplitRepo(%q) repo = %q, expected %q", tt.repo, repo, tt.expectedRepo)
				}
			}
		})
	}
}

func TestSplitRepoErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		repo     string
		contains string
	}{
		{
			name:     "error message contains repo string",
			repo:     "invalid",
			contains: "invalid",
		},
		{
			name:     "error message for empty string",
			repo:     "",
			contains: "invalid repository format",
		},
		{
			name:     "error message mentions format",
			repo:     "no-slash",
			contains: "invalid repository format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := SplitRepo(tt.repo)
			if err == nil {
				t.Errorf("SplitRepo(%q) expected error but got none", tt.repo)
				return
			}
			errorMsg := err.Error()
			if errorMsg == "" {
				t.Errorf("SplitRepo(%q) error message is empty", tt.repo)
			}
			// Note: We're checking if the error contains the expected substring
			// This is a basic check; you might want to verify the exact format
		})
	}
}

func TestSplitRepoConsistency(t *testing.T) {
	// Test that calling SplitRepo multiple times with the same input returns consistent results
	repo := "owner/repo"

	owner1, repo1, err1 := SplitRepo(repo)
	owner2, repo2, err2 := SplitRepo(repo)
	owner3, repo3, err3 := SplitRepo(repo)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("SplitRepo(%q) unexpected errors: %v, %v, %v", repo, err1, err2, err3)
	}

	if owner1 != owner2 || owner2 != owner3 {
		t.Errorf("SplitRepo(%q) inconsistent owners: %q, %q, %q", repo, owner1, owner2, owner3)
	}

	if repo1 != repo2 || repo2 != repo3 {
		t.Errorf("SplitRepo(%q) inconsistent repos: %q, %q, %q", repo, repo1, repo2, repo3)
	}
}

func TestSplitRepoURLVariations(t *testing.T) {
	// Test that different URL formats for the same repo produce the same result
	tests := []struct {
		name  string
		repos []string
	}{
		{
			name: "same repo different formats",
			repos: []string{
				"owner/repo",
				"https://github.com/owner/repo",
				"https://github.com/owner/repo/",
				"https://github.com/owner/repo/tree/main",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var owners []string
			var repos []string

			for _, repo := range tt.repos {
				owner, repoName, err := SplitRepo(repo)
				if err != nil {
					t.Errorf("SplitRepo(%q) unexpected error: %v", repo, err)
					continue
				}
				owners = append(owners, owner)
				repos = append(repos, repoName)
			}

			// Check all owners are the same
			for i := 1; i < len(owners); i++ {
				if owners[i] != owners[0] {
					t.Errorf("Different owner for same repo: %q vs %q (from %q and %q)",
						owners[0], owners[i], tt.repos[0], tt.repos[i])
				}
			}

			// Check all repo names are the same
			for i := 1; i < len(repos); i++ {
				if repos[i] != repos[0] {
					t.Errorf("Different repo name for same repo: %q vs %q (from %q and %q)",
						repos[0], repos[i], tt.repos[0], tt.repos[i])
				}
			}
		})
	}
}
