package config

import "os"

const GITHUB_APP_ID = "GITHUB_APP_ID"
const GITHUB_INSTALLATION_ID = "GITHUB_INSTALLATION_ID"
const GITHUB_PRIVATE_KEY_PATH = "GITHUB_PRIVATE_KEY_PATH"

func UsePrivateKeyAuth() bool {
	appIDStr := os.Getenv(GITHUB_APP_ID)
	installationIDStr := os.Getenv(GITHUB_INSTALLATION_ID)
	privateKeyPath := os.Getenv(GITHUB_PRIVATE_KEY_PATH)

	return appIDStr != "" && installationIDStr != "" && privateKeyPath != ""
}
