package collections

import (
	"fmt"
	"slices"
	"strings"
)

func MapToString(m map[string]string) string {
	var sb strings.Builder

	// process keys in sorted order
	keys := []string{}
	for k, _ := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s;", k, m[k]))
	}

	return sb.String()
}

func ConvertStringMap(input map[string]interface{}) map[string]string {
	if input == nil {
		return nil
	}

	result := make(map[string]string)
	for key, value := range input {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}

func GetOtherValues(input map[string]interface{}, excludeKeys []string) map[string]string {
	result := make(map[string]string)
	excludeMap := make(map[string]bool)
	for _, key := range excludeKeys {
		excludeMap[key] = true
	}

	for key, value := range input {
		if !excludeMap[key] {
			result[key] = fmt.Sprintf("%v", value)
		}
	}

	return result
}

func GetStringProperty(input map[string]interface{}, key string) string {
	usesInterface, ok := input[key]
	if ok {
		usesString, ok := usesInterface.(string)
		if ok {
			// Script steps often don't have a 'uses' field
			return usesString
		}
	}

	return ""
}

func GetChildMap(input map[string]interface{}, key string) map[string]interface{} {
	childInterface, ok := input[key]
	if !ok {
		return nil
	}
	childMap, ok := childInterface.(map[string]interface{})
	if !ok {
		return nil
	}
	return childMap
}
