package configuration

import "os"

func GetEncryptionKey() string {
	return os.Getenv("DUPCOST_ENCRYPTION_KEY")
}
