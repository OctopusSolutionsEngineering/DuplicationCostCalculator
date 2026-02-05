package configuration

import "os"

func GetClientId() string {
	return os.Getenv("DUPCOST_GITHUB_CLIENT_ID")
}
