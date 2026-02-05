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

func TestConvertStringMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			name:     "nil map",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]string{},
		},
		{
			name: "string values",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "integer values",
			input: map[string]interface{}{
				"age":   25,
				"count": 100,
			},
			expected: map[string]string{
				"age":   "25",
				"count": "100",
			},
		},
		{
			name: "float values",
			input: map[string]interface{}{
				"price":       19.99,
				"temperature": -5.5,
			},
			expected: map[string]string{
				"price":       "19.99",
				"temperature": "-5.5",
			},
		},
		{
			name: "boolean values",
			input: map[string]interface{}{
				"enabled":  true,
				"disabled": false,
			},
			expected: map[string]string{
				"enabled":  "true",
				"disabled": "false",
			},
		},
		{
			name: "mixed types",
			input: map[string]interface{}{
				"name":    "John",
				"age":     30,
				"height":  5.9,
				"active":  true,
				"balance": -100.50,
			},
			expected: map[string]string{
				"name":    "John",
				"age":     "30",
				"height":  "5.9",
				"active":  "true",
				"balance": "-100.5",
			},
		},
		{
			name: "nil values",
			input: map[string]interface{}{
				"nullable": nil,
				"string":   "value",
			},
			expected: map[string]string{
				"nullable": "<nil>",
				"string":   "value",
			},
		},
		{
			name: "empty string value",
			input: map[string]interface{}{
				"empty": "",
				"key":   "value",
			},
			expected: map[string]string{
				"empty": "",
				"key":   "value",
			},
		},
		{
			name: "special characters in values",
			input: map[string]interface{}{
				"url":     "https://example.com?param=value&other=123",
				"path":    "/home/user/file.txt",
				"message": "Hello, World!",
			},
			expected: map[string]string{
				"url":     "https://example.com?param=value&other=123",
				"path":    "/home/user/file.txt",
				"message": "Hello, World!",
			},
		},
		{
			name: "zero values",
			input: map[string]interface{}{
				"zero_int":   0,
				"zero_float": 0.0,
				"zero_bool":  false,
			},
			expected: map[string]string{
				"zero_int":   "0",
				"zero_float": "0",
				"zero_bool":  "false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertStringMap(tt.input)

			// Check if both are nil
			if tt.expected == nil && result == nil {
				return
			}

			// Check if one is nil and the other is not
			if (tt.expected == nil) != (result == nil) {
				t.Errorf("ConvertStringMap() = %v, expected %v", result, tt.expected)
				return
			}

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("ConvertStringMap() returned map with %d entries, expected %d", len(result), len(tt.expected))
				return
			}

			// Check each key-value pair
			for key, expectedValue := range tt.expected {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("ConvertStringMap() missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("ConvertStringMap()[%q] = %q, expected %q", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestConvertStringMapPreservesAllKeys(t *testing.T) {
	input := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
		"e": 5,
	}

	result := ConvertStringMap(input)

	if len(result) != len(input) {
		t.Errorf("ConvertStringMap() preserved %d keys, expected %d", len(result), len(input))
	}

	for key := range input {
		if _, exists := result[key]; !exists {
			t.Errorf("ConvertStringMap() lost key %q", key)
		}
	}
}

func TestGetOtherValues(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]interface{}
		excludeKeys []string
		expected    map[string]string
	}{
		{
			name: "exclude no keys",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			excludeKeys: []string{},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "exclude one key",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			excludeKeys: []string{"key2"},
			expected: map[string]string{
				"key1": "value1",
				"key3": "value3",
			},
		},
		{
			name: "exclude multiple keys",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			},
			excludeKeys: []string{"key1", "key3"},
			expected: map[string]string{
				"key2": "value2",
				"key4": "value4",
			},
		},
		{
			name: "exclude all keys",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			excludeKeys: []string{"key1", "key2"},
			expected:    map[string]string{},
		},
		{
			name: "exclude non-existent keys",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			excludeKeys: []string{"key3", "key4"},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:        "empty input map",
			input:       map[string]interface{}{},
			excludeKeys: []string{"key1"},
			expected:    map[string]string{},
		},
		{
			name: "nil exclude keys",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			excludeKeys: nil,
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "mixed value types",
			input: map[string]interface{}{
				"name":   "John",
				"age":    30,
				"height": 5.9,
				"active": true,
			},
			excludeKeys: []string{"age"},
			expected: map[string]string{
				"name":   "John",
				"height": "5.9",
				"active": "true",
			},
		},
		{
			name: "exclude with different casing",
			input: map[string]interface{}{
				"Key1": "value1",
				"key2": "value2",
			},
			excludeKeys: []string{"key1"}, // lowercase
			expected: map[string]string{
				"Key1": "value1", // Key1 is not excluded (case-sensitive)
				"key2": "value2",
			},
		},
		{
			name: "duplicate keys in exclude list",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			excludeKeys: []string{"key1", "key1", "key2", "key2"},
			expected: map[string]string{
				"key3": "value3",
			},
		},
		{
			name: "exclude keys with special characters",
			input: map[string]interface{}{
				"key-1": "value1",
				"key_2": "value2",
				"key.3": "value3",
				"key@4": "value4",
				"key:5": "value5",
			},
			excludeKeys: []string{"key-1", "key.3", "key:5"},
			expected: map[string]string{
				"key_2": "value2",
				"key@4": "value4",
			},
		},
		{
			name: "nil values in input",
			input: map[string]interface{}{
				"key1": nil,
				"key2": "value2",
				"key3": nil,
			},
			excludeKeys: []string{"key3"},
			expected: map[string]string{
				"key1": "<nil>",
				"key2": "value2",
			},
		},
		{
			name: "zero values",
			input: map[string]interface{}{
				"zero_int":     0,
				"zero_float":   0.0,
				"zero_bool":    false,
				"empty_string": "",
			},
			excludeKeys: []string{"zero_int"},
			expected: map[string]string{
				"zero_float":   "0",
				"zero_bool":    "false",
				"empty_string": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOtherValues(tt.input, tt.excludeKeys)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("GetOtherValues() returned map with %d entries, expected %d", len(result), len(tt.expected))
				t.Logf("Result: %v", result)
				t.Logf("Expected: %v", tt.expected)
				return
			}

			// Check each key-value pair
			for key, expectedValue := range tt.expected {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("GetOtherValues() missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("GetOtherValues()[%q] = %q, expected %q", key, actualValue, expectedValue)
				}
			}

			// Check that excluded keys are not present
			for _, excludeKey := range tt.excludeKeys {
				if _, exists := result[excludeKey]; exists {
					t.Errorf("GetOtherValues() should have excluded key %q but it's present with value %q", excludeKey, result[excludeKey])
				}
			}
		})
	}
}

func TestGetOtherValuesDoesNotModifyOriginal(t *testing.T) {
	original := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// Make a copy to compare later
	originalCopy := make(map[string]interface{})
	for k, v := range original {
		originalCopy[k] = v
	}

	excludeKeys := []string{"key2"}

	result := GetOtherValues(original, excludeKeys)

	// Verify the original map was not modified
	if len(original) != len(originalCopy) {
		t.Errorf("GetOtherValues() modified the original map length: got %d, expected %d", len(original), len(originalCopy))
	}

	for key, value := range originalCopy {
		if originalValue, ok := original[key]; !ok {
			t.Errorf("GetOtherValues() removed key %q from original map", key)
		} else if originalValue != value {
			t.Errorf("GetOtherValues() modified value for key %q in original map: got %v, expected %v", key, originalValue, value)
		}
	}

	// Verify result does not contain excluded key
	if _, exists := result["key2"]; exists {
		t.Errorf("GetOtherValues() result should not contain excluded key 'key2'")
	}

	// Verify result contains non-excluded keys
	if result["key1"] != "value1" || result["key3"] != "value3" {
		t.Errorf("GetOtherValues() result missing expected keys")
	}
}

func TestGetStringProperty(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		key      string
		expected string
	}{
		{
			name: "string value exists",
			input: map[string]interface{}{
				"uses": "actions/checkout@v4",
			},
			key:      "uses",
			expected: "actions/checkout@v4",
		},
		{
			name: "key does not exist",
			input: map[string]interface{}{
				"other": "value",
			},
			key:      "uses",
			expected: "",
		},
		{
			name: "value is not a string - integer",
			input: map[string]interface{}{
				"uses": 123,
			},
			key:      "uses",
			expected: "",
		},
		{
			name: "value is not a string - boolean",
			input: map[string]interface{}{
				"uses": true,
			},
			key:      "uses",
			expected: "",
		},
		{
			name: "value is not a string - map",
			input: map[string]interface{}{
				"uses": map[string]interface{}{"nested": "value"},
			},
			key:      "uses",
			expected: "",
		},
		{
			name: "value is not a string - slice",
			input: map[string]interface{}{
				"uses": []string{"item1", "item2"},
			},
			key:      "uses",
			expected: "",
		},
		{
			name: "value is nil",
			input: map[string]interface{}{
				"uses": nil,
			},
			key:      "uses",
			expected: "",
		},
		{
			name:     "empty input map",
			input:    map[string]interface{}{},
			key:      "uses",
			expected: "",
		},
		{
			name:     "nil input map",
			input:    nil,
			key:      "uses",
			expected: "",
		},
		{
			name: "empty string value",
			input: map[string]interface{}{
				"uses": "",
			},
			key:      "uses",
			expected: "",
		},
		{
			name: "string with whitespace",
			input: map[string]interface{}{
				"uses": "  actions/checkout@v3  ",
			},
			key:      "uses",
			expected: "  actions/checkout@v3  ",
		},
		{
			name: "string with special characters",
			input: map[string]interface{}{
				"uses": "docker/build-push-action@v5.0.0",
			},
			key:      "uses",
			expected: "docker/build-push-action@v5.0.0",
		},
		{
			name: "multiple keys, get specific one",
			input: map[string]interface{}{
				"name": "Checkout code",
				"uses": "actions/checkout@v4",
				"with": map[string]interface{}{"fetch-depth": 0},
			},
			key:      "uses",
			expected: "actions/checkout@v4",
		},
		{
			name: "empty key",
			input: map[string]interface{}{
				"":     "value",
				"uses": "actions/checkout@v4",
			},
			key:      "",
			expected: "value",
		},
		{
			name: "value is float (not string)",
			input: map[string]interface{}{
				"uses": 3.14,
			},
			key:      "uses",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringProperty(tt.input, tt.key)
			if result != tt.expected {
				t.Errorf("GetStringProperty(%v, %q) = %q, expected %q", tt.input, tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetChildMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		key      string
		expected map[string]interface{}
	}{
		{
			name: "child map exists",
			input: map[string]interface{}{
				"with": map[string]interface{}{
					"fetch-depth": 0,
					"token":       "${{ secrets.GITHUB_TOKEN }}",
				},
			},
			key: "with",
			expected: map[string]interface{}{
				"fetch-depth": 0,
				"token":       "${{ secrets.GITHUB_TOKEN }}",
			},
		},
		{
			name: "key does not exist",
			input: map[string]interface{}{
				"other": "value",
			},
			key:      "with",
			expected: nil,
		},
		{
			name: "value is not a map - string",
			input: map[string]interface{}{
				"with": "not a map",
			},
			key:      "with",
			expected: nil,
		},
		{
			name: "value is not a map - integer",
			input: map[string]interface{}{
				"with": 123,
			},
			key:      "with",
			expected: nil,
		},
		{
			name: "value is not a map - boolean",
			input: map[string]interface{}{
				"with": true,
			},
			key:      "with",
			expected: nil,
		},
		{
			name: "value is not a map - slice",
			input: map[string]interface{}{
				"with": []string{"item1", "item2"},
			},
			key:      "with",
			expected: nil,
		},
		{
			name: "value is nil",
			input: map[string]interface{}{
				"with": nil,
			},
			key:      "with",
			expected: nil,
		},
		{
			name:     "empty input map",
			input:    map[string]interface{}{},
			key:      "with",
			expected: nil,
		},
		{
			name:     "nil input map",
			input:    nil,
			key:      "with",
			expected: nil,
		},
		{
			name: "empty child map",
			input: map[string]interface{}{
				"with": map[string]interface{}{},
			},
			key:      "with",
			expected: map[string]interface{}{},
		},
		{
			name: "nested maps",
			input: map[string]interface{}{
				"with": map[string]interface{}{
					"nested": map[string]interface{}{
						"deep": "value",
					},
				},
			},
			key: "with",
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"deep": "value",
				},
			},
		},
		{
			name: "child map with mixed types",
			input: map[string]interface{}{
				"with": map[string]interface{}{
					"string_val": "text",
					"int_val":    42,
					"bool_val":   true,
					"float_val":  3.14,
					"nil_val":    nil,
				},
			},
			key: "with",
			expected: map[string]interface{}{
				"string_val": "text",
				"int_val":    42,
				"bool_val":   true,
				"float_val":  3.14,
				"nil_val":    nil,
			},
		},
		{
			name: "multiple keys, get specific child map",
			input: map[string]interface{}{
				"name": "Setup Node",
				"uses": "actions/setup-node@v3",
				"with": map[string]interface{}{
					"node-version": "18",
				},
			},
			key: "with",
			expected: map[string]interface{}{
				"node-version": "18",
			},
		},
		{
			name: "empty key",
			input: map[string]interface{}{
				"": map[string]interface{}{
					"value": "test",
				},
			},
			key: "",
			expected: map[string]interface{}{
				"value": "test",
			},
		},
		{
			name: "child map with special keys",
			input: map[string]interface{}{
				"env": map[string]interface{}{
					"NODE_ENV":    "production",
					"API-KEY":     "secret",
					"config.file": "app.json",
					"var_name":    "value",
				},
			},
			key: "env",
			expected: map[string]interface{}{
				"NODE_ENV":    "production",
				"API-KEY":     "secret",
				"config.file": "app.json",
				"var_name":    "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetChildMap(tt.input, tt.key)

			// Check if both are nil
			if tt.expected == nil && result == nil {
				return
			}

			// Check if one is nil and the other is not
			if (tt.expected == nil) != (result == nil) {
				t.Errorf("GetChildMap(%v, %q) = %v, expected %v", tt.input, tt.key, result, tt.expected)
				return
			}

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("GetChildMap(%v, %q) returned map with %d entries, expected %d", tt.input, tt.key, len(result), len(tt.expected))
				return
			}

			// Check each key-value pair
			for key, expectedValue := range tt.expected {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("GetChildMap(%v, %q) missing key %q", tt.input, tt.key, key)
				} else {
					// Deep comparison for nested maps
					if expectedMap, isMap := expectedValue.(map[string]interface{}); isMap {
						actualMap, ok := actualValue.(map[string]interface{})
						if !ok {
							t.Errorf("GetChildMap(%v, %q)[%q] is not a map", tt.input, tt.key, key)
							continue
						}
						if len(actualMap) != len(expectedMap) {
							t.Errorf("GetChildMap(%v, %q)[%q] has %d entries, expected %d", tt.input, tt.key, key, len(actualMap), len(expectedMap))
						}
						for nestedKey, nestedExpectedValue := range expectedMap {
							if nestedActualValue, ok := actualMap[nestedKey]; !ok {
								t.Errorf("GetChildMap(%v, %q)[%q][%q] missing", tt.input, tt.key, key, nestedKey)
							} else if nestedActualValue != nestedExpectedValue {
								t.Errorf("GetChildMap(%v, %q)[%q][%q] = %v, expected %v", tt.input, tt.key, key, nestedKey, nestedActualValue, nestedExpectedValue)
							}
						}
					} else if actualValue != expectedValue {
						t.Errorf("GetChildMap(%v, %q)[%q] = %v, expected %v", tt.input, tt.key, key, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestGetChildMapReturnsReference(t *testing.T) {
	original := map[string]interface{}{
		"with": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}

	result := GetChildMap(original, "with")

	// Verify we got a result
	if result == nil {
		t.Fatal("GetChildMap() returned nil unexpectedly")
	}

	// Verify it's the same reference (modifying result affects original)
	result["key3"] = "value3"
	result["key1"] = "modified"

	// The original should be modified because maps are reference types in Go
	originalWith := original["with"].(map[string]interface{})
	if len(originalWith) != 3 {
		t.Errorf("GetChildMap() does not return a reference: original map length = %d, expected 3", len(originalWith))
	}

	if originalWith["key1"] != "modified" {
		t.Errorf("GetChildMap() does not return a reference: key1 = %v, expected 'modified'", originalWith["key1"])
	}

	if originalWith["key3"] != "value3" {
		t.Errorf("GetChildMap() does not return a reference: key3 = %v, expected 'value3'", originalWith["key3"])
	}
}
