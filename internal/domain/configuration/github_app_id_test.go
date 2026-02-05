package configuration

import (
	"os"
	"testing"
)

func TestGetGithubAppId(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "valid GitHub App ID",
			envValue: "123456",
			expected: "123456",
		},
		{
			name:     "empty GitHub App ID",
			envValue: "",
			expected: "",
		},
		{
			name:     "long GitHub App ID",
			envValue: "9876543210",
			expected: "9876543210",
		},
		{
			name:     "single digit App ID",
			envValue: "1",
			expected: "1",
		},
		{
			name:     "App ID with leading zeros",
			envValue: "000123",
			expected: "000123",
		},
		{
			name:     "alphanumeric App ID (edge case)",
			envValue: "app123id",
			expected: "app123id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv("GITHUB_APP_ID", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_APP_ID")

			// Call GetGithubAppId
			result := GetGithubAppId()

			// Verify result
			if result != tt.expected {
				t.Errorf("GetGithubAppId() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetGithubAppIdUnset(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("GITHUB_APP_ID")

	result := GetGithubAppId()

	if result != "" {
		t.Errorf("GetGithubAppId() = %q, expected empty string when env var is unset", result)
	}
}

func TestGetGithubAppIdConsistency(t *testing.T) {
	// Test that calling GetGithubAppId multiple times returns consistent results
	testValue := "654321"
	err := os.Setenv("GITHUB_APP_ID", testValue)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("GITHUB_APP_ID")

	results := []string{}
	for i := 0; i < 3; i++ {
		results = append(results, GetGithubAppId())
	}

	// All results should be the same
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Call %d: GetGithubAppId() = %q, expected %q (inconsistent results)", i+1, results[i], results[0])
		}
	}

	if results[0] != testValue {
		t.Errorf("GetGithubAppId() = %q, expected %q", results[0], testValue)
	}
}

func TestGetGithubAppIdEnvironmentVariableName(t *testing.T) {
	// Test that the function reads from the correct environment variable name
	correctEnvVar := "GITHUB_APP_ID"
	incorrectEnvVar := "GITHUB_APPLICATION_ID"

	// Set incorrect env var
	os.Setenv(incorrectEnvVar, "999999")
	defer os.Unsetenv(incorrectEnvVar)

	// Set correct env var
	expectedValue := "123456"
	os.Setenv(correctEnvVar, expectedValue)
	defer os.Unsetenv(correctEnvVar)

	result := GetGithubAppId()

	if result != expectedValue {
		t.Errorf("GetGithubAppId() = %q, expected %q (should read from %s, not %s)",
			result, expectedValue, correctEnvVar, incorrectEnvVar)
	}

	if result == "999999" {
		t.Error("GetGithubAppId() read from wrong environment variable")
	}
}

func TestGetGithubAppIdWithWhitespace(t *testing.T) {
	// Test that whitespace is preserved (not trimmed)
	// Note: In practice, App IDs shouldn't have whitespace, but this tests the actual behavior
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "leading whitespace",
			envValue: "  123456",
			expected: "  123456",
		},
		{
			name:     "trailing whitespace",
			envValue: "123456  ",
			expected: "123456  ",
		},
		{
			name:     "both leading and trailing whitespace",
			envValue: "  123456  ",
			expected: "  123456  ",
		},
		{
			name:     "whitespace only",
			envValue: "   ",
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("GITHUB_APP_ID", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_APP_ID")

			result := GetGithubAppId()

			if result != tt.expected {
				t.Errorf("GetGithubAppId() = %q, expected %q (whitespace not preserved as expected)", result, tt.expected)
			}
		})
	}
}

func TestGetGithubAppIdRealWorldFormats(t *testing.T) {
	// Test with realistic GitHub App ID formats
	tests := []struct {
		name     string
		envValue string
	}{
		{
			name:     "typical 6-digit App ID",
			envValue: "123456",
		},
		{
			name:     "typical 7-digit App ID",
			envValue: "1234567",
		},
		{
			name:     "typical 8-digit App ID",
			envValue: "12345678",
		},
		{
			name:     "small App ID",
			envValue: "1",
		},
		{
			name:     "large App ID",
			envValue: "999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("GITHUB_APP_ID", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("GITHUB_APP_ID")

			result := GetGithubAppId()

			if result != tt.envValue {
				t.Errorf("GetGithubAppId() = %q, expected %q", result, tt.envValue)
			}

			// Verify no transformation occurred
			if len(result) > 0 && result != tt.envValue {
				t.Error("GetGithubAppId() modified the App ID, which could indicate an issue")
			}
		})
	}
}
