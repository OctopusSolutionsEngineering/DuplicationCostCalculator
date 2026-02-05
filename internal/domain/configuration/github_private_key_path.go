package configuration

import "os"

const GITHUB_PRIVATE_KEY_PATH = "GITHUB_PRIVATE_KEY_PATH"

func GetGitHubPrivateKeyPath() string {
	return os.Getenv(GITHUB_PRIVATE_KEY_PATH)
}
