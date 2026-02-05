package configuration

import (
	"os"
	"testing"
)

func TestGetGitHubInstallationId(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "valid GitHub Installation ID",
			envValue: "12345678",
			expected: "12345678",
		},
		{
			name:     "empty GitHub Installation ID",
			envValue: "",
			expected: "",
		},
		{
			name:     "long GitHub Installation ID",
			envValue: "98765432109876",
			expected: "98765432109876",
		},
		{
			name:     "single digit Installation ID",
			envValue: "1",
			expected: "1",
		},
		{
			name:     "Installation ID with leading zeros",
			envValue: "000123",
			expected: "000123",
		},
		{
			name:     "typical 8-digit Installation ID",
			envValue: "87654321",
			expected: "87654321",
		},
		{
			name:     "alphanumeric Installation ID (edge case)",
			envValue: "install123id",
			expected: "install123id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv("GITHUB_INSTALLATION_ID", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_INSTALLATION_ID")

			// Call GetGitHubInstallationId
			result := GetGitHubInstallationId()

			// Verify result
			if result != tt.expected {
				t.Errorf("GetGitHubInstallationId() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetGitHubInstallationIdUnset(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("GITHUB_INSTALLATION_ID")

	result := GetGitHubInstallationId()

	if result != "" {
		t.Errorf("GetGitHubInstallationId() = %q, expected empty string when env var is unset", result)
	}
}

func TestGetGitHubInstallationIdConsistency(t *testing.T) {
	// Test that calling GetGitHubInstallationId multiple times returns consistent results
	testValue := "11223344"
	err := os.Setenv("GITHUB_INSTALLATION_ID", testValue)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("GITHUB_INSTALLATION_ID")

	results := []string{}
	for i := 0; i < 3; i++ {
		results = append(results, GetGitHubInstallationId())
	}

	// All results should be the same
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Call %d: GetGitHubInstallationId() = %q, expected %q (inconsistent results)", i+1, results[i], results[0])
		}
	}

	if results[0] != testValue {
		t.Errorf("GetGitHubInstallationId() = %q, expected %q", results[0], testValue)
	}
}

func TestGetGitHubInstallationIdEnvironmentVariableName(t *testing.T) {
	// Test that the function reads from the correct environment variable name
	correctEnvVar := "GITHUB_INSTALLATION_ID"
	incorrectEnvVar := "GITHUB_INSTALL_ID"

	// Set incorrect env var
	os.Setenv(incorrectEnvVar, "99999999")
	defer os.Unsetenv(incorrectEnvVar)

	// Set correct env var
	expectedValue := "12345678"
	os.Setenv(correctEnvVar, expectedValue)
	defer os.Unsetenv(correctEnvVar)

	result := GetGitHubInstallationId()

	if result != expectedValue {
		t.Errorf("GetGitHubInstallationId() = %q, expected %q (should read from %s, not %s)",
			result, expectedValue, correctEnvVar, incorrectEnvVar)
	}

	if result == "99999999" {
		t.Error("GetGitHubInstallationId() read from wrong environment variable")
	}
}

func TestGetGitHubInstallationIdWithWhitespace(t *testing.T) {
	// Test that whitespace is preserved (not trimmed)
	// Note: In practice, Installation IDs shouldn't have whitespace, but this tests the actual behavior
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "leading whitespace",
			envValue: "  12345678",
			expected: "  12345678",
		},
		{
			name:     "trailing whitespace",
			envValue: "12345678  ",
			expected: "12345678  ",
		},
		{
			name:     "both leading and trailing whitespace",
			envValue: "  12345678  ",
			expected: "  12345678  ",
		},
		{
			name:     "whitespace only",
			envValue: "   ",
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("GITHUB_INSTALLATION_ID", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_INSTALLATION_ID")

			result := GetGitHubInstallationId()

			if result != tt.expected {
				t.Errorf("GetGitHubInstallationId() = %q, expected %q (whitespace not preserved as expected)", result, tt.expected)
			}
		})
	}
}

func TestGetGitHubInstallationIdRealWorldFormats(t *testing.T) {
	// Test with realistic GitHub Installation ID formats
	tests := []struct {
		name     string
		envValue string
	}{
		{
			name:     "typical 8-digit Installation ID",
			envValue: "12345678",
		},
		{
			name:     "typical 9-digit Installation ID",
			envValue: "123456789",
		},
		{
			name:     "typical 10-digit Installation ID",
			envValue: "1234567890",
		},
		{
			name:     "small Installation ID",
			envValue: "1",
		},
		{
			name:     "large Installation ID",
			envValue: "9999999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("GITHUB_INSTALLATION_ID", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_INSTALLATION_ID")

			result := GetGitHubInstallationId()

			if result != tt.envValue {
				t.Errorf("GetGitHubInstallationId() = %q, expected %q", result, tt.envValue)
			}

			// Verify no transformation occurred
			if len(result) > 0 && result != tt.envValue {
				t.Error("GetGitHubInstallationId() modified the Installation ID, which could indicate an issue")
			}
		})
	}
}
