package parsing

import "strings"

func GetActionIdAndVersion(uses string) (string, string) {
	if strings.Contains(uses, "@") {
		parts := strings.SplitN(uses, "@", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	} else {
		return uses, "latest"
	}
}
