package parsing

import (
	"fmt"
	"strings"
)

func SplitRepo(repo string) (string, string, error) {
	// Split repo into owner and name
	parts := strings.Split(SanitizeRepo(repo), "/")
	// Ignore any paths that may have been on the end of a url
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository format: %s", repo)
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format: %s", repo)
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func SplitRepoNoErr(repo string) (string, string) {
	owner, repoName, err := SplitRepo(repo)
	if err != nil {
		return "", ""
	}

	return owner, repoName
}

func SanitizeRepo(repo string) string {
	value := strings.Replace(repo, "https://github.com/", "", 1)
	value = strings.Replace(value, "github.com/", "", 1)
	value = strings.Replace(value, ".git", "", 1)
	return strings.TrimSpace(value)
}
