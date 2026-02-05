package configuration

import "os"

func GetClientSecret() string {
	return os.Getenv("DUPCOST_GITHUB_CLIENT_SECRET")
}
