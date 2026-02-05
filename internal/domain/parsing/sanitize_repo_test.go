package parsing

import (
	"testing"
)

func TestSanitizeRepo(t *testing.T) {
	tests := []struct {
		name     string
		repo     string
		expected string
	}{
		{
			name:     "owner/repo format unchanged",
			repo:     "owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "https URL with github.com",
			repo:     "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "github.com without https",
			repo:     "github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "https URL with trailing path",
			repo:     "https://github.com/owner/repo/tree/main",
			expected: "owner/repo/tree/main",
		},
		{
			name:     "https URL with issues path",
			repo:     "https://github.com/owner/repo/issues/123",
			expected: "owner/repo/issues/123",
		},
		{
			name:     "github.com without https with trailing path",
			repo:     "github.com/owner/repo/pulls",
			expected: "owner/repo/pulls",
		},
		{
			name:     "owner/repo with hyphens",
			repo:     "my-org/my-repo",
			expected: "my-org/my-repo",
		},
		{
			name:     "owner/repo with underscores",
			repo:     "my_org/my_repo",
			expected: "my_org/my_repo",
		},
		{
			name:     "owner/repo with dots",
			repo:     "my.org/my.repo",
			expected: "my.org/my.repo",
		},
		{
			name:     "owner/repo with numbers",
			repo:     "org123/repo456",
			expected: "org123/repo456",
		},
		{
			name:     "https URL with .git suffix",
			repo:     "https://github.com/owner/repo.git",
			expected: "owner/repo.git",
		},
		{
			name:     "empty string",
			repo:     "",
			expected: "",
		},
		{
			name:     "only owner",
			repo:     "owner",
			expected: "owner",
		},
		{
			name:     "only slash",
			repo:     "/",
			expected: "/",
		},
		{
			name:     "real world example - OctopusDeploy",
			repo:     "https://github.com/OctopusDeploy/OctopusDeploy",
			expected: "OctopusDeploy/OctopusDeploy",
		},
		{
			name:     "real world example - actions/checkout",
			repo:     "https://github.com/actions/checkout",
			expected: "actions/checkout",
		},
		{
			name:     "case sensitivity preserved",
			repo:     "https://github.com/MyOrg/MyRepo",
			expected: "MyOrg/MyRepo",
		},
		{
			name:     "multiple slashes",
			repo:     "owner/repo/extra/parts",
			expected: "owner/repo/extra/parts",
		},
		{
			name:     "URL with query parameters",
			repo:     "https://github.com/owner/repo?tab=readme",
			expected: "owner/repo?tab=readme",
		},
		{
			name:     "URL with fragment",
			repo:     "https://github.com/owner/repo#readme",
			expected: "owner/repo#readme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeRepo(tt.repo)
			if result != tt.expected {
				t.Errorf("SanitizeRepo(%q) = %q, expected %q", tt.repo, result, tt.expected)
			}
		})
	}
}

func TestSanitizeRepoConsistency(t *testing.T) {
	// Test that calling SanitizeRepo multiple times produces consistent results
	testCases := []string{
		"owner/repo",
		"https://github.com/owner/repo",
		"github.com/owner/repo",
	}

	for _, tc := range testCases {
		result1 := SanitizeRepo(tc)
		result2 := SanitizeRepo(tc)
		result3 := SanitizeRepo(tc)

		if result1 != result2 || result2 != result3 {
			t.Errorf("SanitizeRepo(%q) produced inconsistent results: %q, %q, %q", tc, result1, result2, result3)
		}
	}
}

func TestSanitizeRepoURLVariations(t *testing.T) {
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
				"github.com/owner/repo",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []string

			for _, repo := range tt.repos {
				results = append(results, SanitizeRepo(repo))
			}

			// All results should be the same
			for i := 1; i < len(results); i++ {
				if results[i] != results[0] {
					t.Errorf("Different result for same repo: %q vs %q (from %q and %q)",
						results[0], results[i], tt.repos[0], tt.repos[i])
				}
			}
		})
	}
}

func TestSanitizeRepoOnlyRemovesPrefix(t *testing.T) {
	// Test that the function removes the first occurrence (with limit 1)
	// Note: strings.Replace will still replace the pattern anywhere it appears first
	tests := []struct {
		name     string
		repo     string
		expected string
	}{
		{
			name:     "https://github.com appears only once",
			repo:     "https://github.com/owner/repo",
			expected: "owner/repo",
		},
		{
			name:     "github.com appears only once",
			repo:     "github.com/owner/repo",
			expected: "owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeRepo(tt.repo)
			if result != tt.expected {
				t.Errorf("SanitizeRepo(%q) = %q, expected %q", tt.repo, result, tt.expected)
			}
		})
	}
}

func TestSanitizeRepoDoesNotTrim(t *testing.T) {
	// Test that whitespace is preserved
	tests := []struct {
		name     string
		repo     string
		expected string
	}{
		{
			name:     "leading whitespace preserved",
			repo:     "  owner/repo",
			expected: "  owner/repo",
		},
		{
			name:     "trailing whitespace preserved",
			repo:     "owner/repo  ",
			expected: "owner/repo  ",
		},
		{
			name:     "both leading and trailing whitespace preserved",
			repo:     "  owner/repo  ",
			expected: "  owner/repo  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeRepo(tt.repo)
			if result != tt.expected {
				t.Errorf("SanitizeRepo(%q) = %q, expected %q", tt.repo, result, tt.expected)
			}
		})
	}
}
