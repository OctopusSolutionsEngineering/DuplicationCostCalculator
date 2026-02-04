package collections

import (
	"testing"
)

func TestMapToString(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: "",
		},
		{
			name: "single entry",
			input: map[string]string{
				"key1": "value1",
			},
			expected: "key1=value1;",
		},
		{
			name: "multiple entries sorted by key",
			input: map[string]string{
				"zebra": "last",
				"alpha": "first",
				"beta":  "second",
			},
			expected: "alpha=first;beta=second;zebra=last;",
		},
		{
			name: "keys with special characters",
			input: map[string]string{
				"key-1": "value-1",
				"key_2": "value_2",
				"key.3": "value.3",
			},
			expected: "key-1=value-1;key.3=value.3;key_2=value_2;",
		},
		{
			name: "values with special characters",
			input: map[string]string{
				"key1": "value with spaces",
				"key2": "value=with=equals",
				"key3": "value;with;semicolons",
			},
			expected: "key1=value with spaces;key2=value=with=equals;key3=value;with;semicolons;",
		},
		{
			name: "empty values",
			input: map[string]string{
				"key1": "",
				"key2": "value2",
				"key3": "",
			},
			expected: "key1=;key2=value2;key3=;",
		},
		{
			name: "numeric-like keys",
			input: map[string]string{
				"3": "three",
				"1": "one",
				"2": "two",
			},
			expected: "1=one;2=two;3=three;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapToString(tt.input)
			if result != tt.expected {
				t.Errorf("MapToString() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMapToStringSortingConsistency(t *testing.T) {
	// Test that the same map produces the same output consistently
	input := map[string]string{
		"z": "26",
		"a": "1",
		"m": "13",
		"b": "2",
	}

	result1 := MapToString(input)
	result2 := MapToString(input)
	result3 := MapToString(input)

	if result1 != result2 || result2 != result3 {
		t.Errorf("MapToString() produces inconsistent results: %q, %q, %q", result1, result2, result3)
	}

	expected := "a=1;b=2;m=13;z=26;"
	if result1 != expected {
		t.Errorf("MapToString() = %q, expected %q", result1, expected)
	}
}
