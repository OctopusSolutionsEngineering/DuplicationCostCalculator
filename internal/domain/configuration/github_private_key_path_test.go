package configuration

import (
	"os"
	"testing"
)

func TestGetGitHubPrivateKeyPath(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "valid absolute path",
			envValue: "/path/to/private-key.pem",
			expected: "/path/to/private-key.pem",
		},
		{
			name:     "valid relative path",
			envValue: "./private-key.pem",
			expected: "./private-key.pem",
		},
		{
			name:     "empty path",
			envValue: "",
			expected: "",
		},
		{
			name:     "path with spaces",
			envValue: "/path/to/my private key.pem",
			expected: "/path/to/my private key.pem",
		},
		{
			name:     "Windows-style path",
			envValue: "C:\\Users\\user\\private-key.pem",
			expected: "C:\\Users\\user\\private-key.pem",
		},
		{
			name:     "path with special characters",
			envValue: "/path/to/key-file_2024.pem",
			expected: "/path/to/key-file_2024.pem",
		},
		{
			name:     "home directory path",
			envValue: "~/secrets/private-key.pem",
			expected: "~/secrets/private-key.pem",
		},
		{
			name:     "path with multiple extensions",
			envValue: "/path/to/key.private.pem",
			expected: "/path/to/key.private.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv("GITHUB_PRIVATE_KEY_PATH", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_PRIVATE_KEY_PATH")

			// Call GetGitHubPrivateKeyPath
			result := GetGitHubPrivateKeyPath()

			// Verify result
			if result != tt.expected {
				t.Errorf("GetGitHubPrivateKeyPath() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetGitHubPrivateKeyPathUnset(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("GITHUB_PRIVATE_KEY_PATH")

	result := GetGitHubPrivateKeyPath()

	if result != "" {
		t.Errorf("GetGitHubPrivateKeyPath() = %q, expected empty string when env var is unset", result)
	}
}

func TestGetGitHubPrivateKeyPathConsistency(t *testing.T) {
	// Test that calling GetGitHubPrivateKeyPath multiple times returns consistent results
	testValue := "/etc/secrets/github-app-key.pem"
	err := os.Setenv("GITHUB_PRIVATE_KEY_PATH", testValue)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("GITHUB_PRIVATE_KEY_PATH")

	results := []string{}
	for i := 0; i < 3; i++ {
		results = append(results, GetGitHubPrivateKeyPath())
	}

	// All results should be the same
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Call %d: GetGitHubPrivateKeyPath() = %q, expected %q (inconsistent results)", i+1, results[i], results[0])
		}
	}

	if results[0] != testValue {
		t.Errorf("GetGitHubPrivateKeyPath() = %q, expected %q", results[0], testValue)
	}
}

func TestGetGitHubPrivateKeyPathEnvironmentVariableName(t *testing.T) {
	// Test that the function reads from the correct environment variable name
	correctEnvVar := "GITHUB_PRIVATE_KEY_PATH"
	incorrectEnvVar := "GITHUB_PRIVATE_KEY"

	// Set incorrect env var
	os.Setenv(incorrectEnvVar, "/wrong/path.pem")
	defer os.Unsetenv(incorrectEnvVar)

	// Set correct env var
	expectedValue := "/correct/path.pem"
	os.Setenv(correctEnvVar, expectedValue)
	defer os.Unsetenv(correctEnvVar)

	result := GetGitHubPrivateKeyPath()

	if result != expectedValue {
		t.Errorf("GetGitHubPrivateKeyPath() = %q, expected %q (should read from %s, not %s)",
			result, expectedValue, correctEnvVar, incorrectEnvVar)
	}

	if result == "/wrong/path.pem" {
		t.Error("GetGitHubPrivateKeyPath() read from wrong environment variable")
	}
}

func TestGetGitHubPrivateKeyPathWithWhitespace(t *testing.T) {
	// Test that whitespace is preserved (not trimmed)
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "leading whitespace",
			envValue: "  /path/to/key.pem",
			expected: "  /path/to/key.pem",
		},
		{
			name:     "trailing whitespace",
			envValue: "/path/to/key.pem  ",
			expected: "/path/to/key.pem  ",
		},
		{
			name:     "both leading and trailing whitespace",
			envValue: "  /path/to/key.pem  ",
			expected: "  /path/to/key.pem  ",
		},
		{
			name:     "whitespace only",
			envValue: "   ",
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("GITHUB_PRIVATE_KEY_PATH", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_PRIVATE_KEY_PATH")

			result := GetGitHubPrivateKeyPath()

			if result != tt.expected {
				t.Errorf("GetGitHubPrivateKeyPath() = %q, expected %q (whitespace not preserved as expected)", result, tt.expected)
			}
		})
	}
}

func TestGetGitHubPrivateKeyPathRealWorldFormats(t *testing.T) {
	// Test with realistic private key path formats
	tests := []struct {
		name     string
		envValue string
	}{
		{
			name:     "typical Linux absolute path",
			envValue: "/etc/secrets/github-app-private-key.pem",
		},
		{
			name:     "typical relative path",
			envValue: "./secrets/private-key.pem",
		},
		{
			name:     "home directory relative path",
			envValue: "~/.ssh/github-app-key.pem",
		},
		{
			name:     "Docker secret path",
			envValue: "/run/secrets/github_app_key",
		},
		{
			name:     "Kubernetes mounted secret",
			envValue: "/var/run/secrets/github/private-key.pem",
		},
		{
			name:     "Windows path",
			envValue: "C:\\ProgramData\\GitHubApp\\private-key.pem",
		},
		{
			name:     "path with environment variable placeholder (not expanded)",
			envValue: "$HOME/.ssh/github-key.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("GITHUB_PRIVATE_KEY_PATH", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_PRIVATE_KEY_PATH")

			result := GetGitHubPrivateKeyPath()

			if result != tt.envValue {
				t.Errorf("GetGitHubPrivateKeyPath() = %q, expected %q", result, tt.envValue)
			}

			// Verify no transformation occurred
			if len(result) > 0 && result != tt.envValue {
				t.Error("GetGitHubPrivateKeyPath() modified the path, which could indicate an issue")
			}
		})
	}
}
