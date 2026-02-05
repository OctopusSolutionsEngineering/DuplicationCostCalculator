package workflows

import "strings"

// HasVersionDrift checks if two actions represent the same action but with different versions.
// It returns true if:
// - Both actions have a non-empty Uses field
// - Both actions have non-empty UsesVersion fields
// - The Uses fields match (same action)
// - The UsesVersion fields differ (different versions)
func HasVersionDrift(action1, action2 Action) bool {
	return action1.Uses != "" &&
		action1.UsesVersion != "" &&
		action2.UsesVersion != "" &&
		action1.Uses == action2.Uses &&
		action1.UsesVersion != action2.UsesVersion
}

func GetActionIdAndVersion(uses string) (string, string) {
	if strings.Contains(uses, "@") {
		parts := strings.SplitN(uses, "@", 2)
		return parts[0], parts[1]
	} else {
		return uses, "latest"
	}
}
