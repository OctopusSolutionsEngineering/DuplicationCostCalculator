package parsing

import (
	"fmt"
	"strings"
)

func SplitRepo(repo string) (string, string, error) {
	// Split repo into owner and name
	parts := strings.Split(strings.Replace(repo, "https://github.com/", "", 1), "/")
	// Ignore any paths that may have been on the end of a url
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository format: %s", repo)
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format: %s", repo)
	}

	return parts[0], parts[1], nil
}
