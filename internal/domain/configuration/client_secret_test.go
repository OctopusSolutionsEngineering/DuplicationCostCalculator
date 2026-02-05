package configuration

import (
	"os"
	"testing"
)

func TestGetClientSecret(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "valid client secret",
			envValue: "abc123def456ghi789jkl012mno345pqr678stu901",
			expected: "abc123def456ghi789jkl012mno345pqr678stu901",
		},
		{
			name:     "empty client secret",
			envValue: "",
			expected: "",
		},
		{
			name:     "client secret with special characters",
			envValue: "secret-with-hyphens_and_underscores",
			expected: "secret-with-hyphens_and_underscores",
		},
		{
			name:     "long client secret",
			envValue: "very-long-client-secret-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
			expected: "very-long-client-secret-1234567890abcdefghijklmnopqrstuvwxyz1234567890",
		},
		{
			name:     "client secret with numbers only",
			envValue: "1234567890",
			expected: "1234567890",
		},
		{
			name:     "client secret with mixed case",
			envValue: "MixedCaseSecret123ABC",
			expected: "MixedCaseSecret123ABC",
		},
		{
			name:     "client secret with symbols",
			envValue: "secret!@#$%^&*()",
			expected: "secret!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			err := os.Setenv("DUPCOST_GITHUB_CLIENT_SECRET", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("DUPCOST_GITHUB_CLIENT_SECRET")

			// Call GetClientSecret
			result := GetClientSecret()

			// Verify result
			if result != tt.expected {
				t.Errorf("GetClientSecret() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetClientSecretUnset(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("DUPCOST_GITHUB_CLIENT_SECRET")

	result := GetClientSecret()

	if result != "" {
		t.Errorf("GetClientSecret() = %q, expected empty string when env var is unset", result)
	}
}

func TestGetClientSecretConsistency(t *testing.T) {
	// Test that calling GetClientSecret multiple times returns consistent results
	testValue := "consistent-client-secret-12345"
	err := os.Setenv("DUPCOST_GITHUB_CLIENT_SECRET", testValue)
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer os.Unsetenv("DUPCOST_GITHUB_CLIENT_SECRET")

	results := []string{}
	for i := 0; i < 3; i++ {
		results = append(results, GetClientSecret())
	}

	// All results should be the same
	for i := 1; i < len(results); i++ {
		if results[i] != results[0] {
			t.Errorf("Call %d: GetClientSecret() = %q, expected %q (inconsistent results)", i+1, results[i], results[0])
		}
	}

	if results[0] != testValue {
		t.Errorf("GetClientSecret() = %q, expected %q", results[0], testValue)
	}
}

func TestGetClientSecretEnvironmentVariableName(t *testing.T) {
	// Test that the function reads from the correct environment variable name
	correctEnvVar := "DUPCOST_GITHUB_CLIENT_SECRET"
	incorrectEnvVar := "GITHUB_CLIENT_SECRET"

	// Set incorrect env var
	os.Setenv(incorrectEnvVar, "should-not-be-read")
	defer os.Unsetenv(incorrectEnvVar)

	// Set correct env var
	expectedValue := "correct-client-secret"
	os.Setenv(correctEnvVar, expectedValue)
	defer os.Unsetenv(correctEnvVar)

	result := GetClientSecret()

	if result != expectedValue {
		t.Errorf("GetClientSecret() = %q, expected %q (should read from %s, not %s)",
			result, expectedValue, correctEnvVar, incorrectEnvVar)
	}

	if result == "should-not-be-read" {
		t.Error("GetClientSecret() read from wrong environment variable")
	}
}

func TestGetClientSecretWithWhitespace(t *testing.T) {
	// Test that whitespace is preserved (not trimmed)
	tests := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "leading whitespace",
			envValue: "  client-secret",
			expected: "  client-secret",
		},
		{
			name:     "trailing whitespace",
			envValue: "client-secret  ",
			expected: "client-secret  ",
		},
		{
			name:     "both leading and trailing whitespace",
			envValue: "  client-secret  ",
			expected: "  client-secret  ",
		},
		{
			name:     "whitespace only",
			envValue: "   ",
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("DUPCOST_GITHUB_CLIENT_SECRET", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("DUPCOST_GITHUB_CLIENT_SECRET")

			result := GetClientSecret()

			if result != tt.expected {
				t.Errorf("GetClientSecret() = %q, expected %q (whitespace not preserved as expected)", result, tt.expected)
			}
		})
	}
}

func TestGetClientSecretSecurityImplications(t *testing.T) {
	// Test that verifies the function works with typical secret patterns
	// This doesn't test actual GitHub secrets, but validates the function handles them
	tests := []struct {
		name     string
		envValue string
	}{
		{
			name:     "hex encoded secret",
			envValue: "abcdef0123456789abcdef0123456789abcdef01",
		},
		{
			name:     "base64-like secret",
			envValue: "ABC123def456GHI789jkl012MNO345pqr678STU901vwx==",
		},
		{
			name:     "alphanumeric secret",
			envValue: "ghp_1234567890abcdefghijklmnopqrstuvwx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv("DUPCOST_GITHUB_CLIENT_SECRET", tt.envValue)
			if err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer os.Unsetenv("DUPCOST_GITHUB_CLIENT_SECRET")

			result := GetClientSecret()

			if result != tt.envValue {
				t.Errorf("GetClientSecret() = %q, expected %q", result, tt.envValue)
			}

			// Ensure the function doesn't accidentally log or expose the secret
			// (this is more of a reminder for developers)
			if len(result) > 0 && result != tt.envValue {
				t.Error("GetClientSecret() returned a modified secret, which could indicate a security issue")
			}
		})
	}
}
