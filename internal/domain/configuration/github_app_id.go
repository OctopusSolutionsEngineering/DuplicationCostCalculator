package configuration

import "os"

const GITHUB_APP_ID = "GITHUB_APP_ID"

func GetGithubAppId() string {
	return os.Getenv(GITHUB_APP_ID)
}
