package workflows

import "github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/models"

// HasVersionDrift checks if two actions represent the same action but with different versions.
// It returns true if:
// - Both actions have a non-empty Uses field
// - Both actions have non-empty UsesVersion fields
// - The Uses fields match (same action)
// - The UsesVersion fields differ (different versions)
func HasVersionDrift(action1, action2 models.Action) bool {
	return action1.Uses != "" &&
		action1.UsesVersion != "" &&
		action2.UsesVersion != "" &&
		action1.Uses == action2.Uses &&
		action1.UsesVersion != action2.UsesVersion
}
