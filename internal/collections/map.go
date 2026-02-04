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
