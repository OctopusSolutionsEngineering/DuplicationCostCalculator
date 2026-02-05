package configuration

import "os"

const GITHUB_INSTALLATION_ID = "GITHUB_INSTALLATION_ID"

func GetGitHubInstallationId() string {
	return os.Getenv(GITHUB_INSTALLATION_ID)
}
