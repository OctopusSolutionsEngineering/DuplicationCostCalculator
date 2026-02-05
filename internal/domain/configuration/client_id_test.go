package configuration

import (
	"os"
	"testing"
)

func TestGetClientId(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "valid client ID",
			envValue: "Iv1.1234567890abcdef",
			expected: "Iv1.1234567890abcdef",
		},
		{
			name:     "empty client ID",
			envValue: "",
			expected: "",
		},
		{
			name:     "client ID with special characters",
			envValue: "client-id-with-hyphens",
			expected: "client-id-with-hyphens",
		},
		{
			name:     "long client ID",
			envValue: "Iv1.1234567890abcdef1234567890abcdef1234567890abcdef",
			expected: "Iv1.1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			name:     "client ID with numbers",
			envValue: "123456789",
			expected: "123456789",
		},
		{
			name:     "client ID with mixed case",
			envValue: "MixedCaseClientId123",
			expected: "MixedCaseClientId123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv("DUPCOST_GITHUB_CLIENT_ID", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("DUPCOST_GITHUB_CLIENT_ID")

			// Call GetClientId
			result := GetClientId()

			// Verify result
			if result != tt.expected {
				t.Errorf("GetClientId() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetClientIdUnset(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("DUPCOST_GITHUB_CLIENT_ID")

	result := GetClientId()

	if result != "" {
		t.Errorf("GetClientId() = %q, expected empty string when env var is unset", result)
	}
}

func TestGetClientIdConsistency(t *testing.T) {
	// Test that calling GetClientId multiple times returns consistent results
	testValue := "consistent-client-id"
	err := os.Setenv("DUPCOST_GITHUB_CLIENT_ID", testValue)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("DUPCOST_GITHUB_CLIENT_ID")

	results := []string{}
	for i := 0; i < 3; i++ {
		results = append(results, GetClientId())
	}

	// All results should be the same
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Call %d: GetClientId() = %q, expected %q (inconsistent results)", i+1, results[i], results[0])
		}
	}

	if results[0] != testValue {
		t.Errorf("GetClientId() = %q, expected %q", results[0], testValue)
	}
}

func TestGetClientIdEnvironmentVariableName(t *testing.T) {
	// Test that the function reads from the correct environment variable name
	correctEnvVar := "DUPCOST_GITHUB_CLIENT_ID"
	incorrectEnvVar := "GITHUB_CLIENT_ID"

	// Set incorrect env var
	os.Setenv(incorrectEnvVar, "should-not-be-read")
	defer os.Unsetenv(incorrectEnvVar)

	// Set correct env var
	expectedValue := "correct-client-id"
	os.Setenv(correctEnvVar, expectedValue)
	defer os.Unsetenv(correctEnvVar)

	result := GetClientId()

	if result != expectedValue {
		t.Errorf("GetClientId() = %q, expected %q (should read from %s, not %s)",
			result, expectedValue, correctEnvVar, incorrectEnvVar)
	}

	if result == "should-not-be-read" {
		t.Error("GetClientId() read from wrong environment variable")
	}
}

func TestGetClientIdWithWhitespace(t *testing.T) {
	// Test that whitespace is preserved (not trimmed)
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "leading whitespace",
			envValue: "  client-id",
			expected: "  client-id",
		},
		{
			name:     "trailing whitespace",
			envValue: "client-id  ",
			expected: "client-id  ",
		},
		{
			name:     "both leading and trailing whitespace",
			envValue: "  client-id  ",
			expected: "  client-id  ",
		},
		{
			name:     "whitespace only",
			envValue: "   ",
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("DUPCOST_GITHUB_CLIENT_ID", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("DUPCOST_GITHUB_CLIENT_ID")

			result := GetClientId()

			if result != tt.expected {
				t.Errorf("GetClientId() = %q, expected %q (whitespace not preserved as expected)", result, tt.expected)
			}
		})
	}
}
