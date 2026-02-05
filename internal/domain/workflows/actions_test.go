package workflows

import (
	"testing"

	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/models"
)

func TestHasVersionDrift(t *testing.T) {
	tests := []struct {
		name     string
		action1  models.Action
		action2  models.Action
		expected bool
	}{
		{
			name: "same action with different versions",
			action1: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v3",
			},
			action2: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v4",
			},
			expected: true,
		},
		{
			name: "same action with same versions",
			action1: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v3",
			},
			action2: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v3",
			},
			expected: false,
		},
		{
			name: "different actions with different versions",
			action1: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v3",
			},
			action2: models.Action{
				Uses:        "actions/setup-node",
				UsesVersion: "v4",
			},
			expected: false,
		},
		{
			name: "action1 with empty Uses",
			action1: models.Action{
				Uses:        "",
				UsesVersion: "v3",
			},
			action2: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v4",
			},
			expected: false,
		},
		{
			name: "action1 with empty UsesVersion",
			action1: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "",
			},
			action2: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v4",
			},
			expected: false,
		},
		{
			name: "action2 with empty UsesVersion",
			action1: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v3",
			},
			action2: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "",
			},
			expected: false,
		},
		{
			name: "both actions with empty Uses",
			action1: models.Action{
				Uses:        "",
				UsesVersion: "v3",
			},
			action2: models.Action{
				Uses:        "",
				UsesVersion: "v4",
			},
			expected: false,
		},
		{
			name: "both actions with empty UsesVersion",
			action1: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "",
			},
			action2: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "",
			},
			expected: false,
		},
		{
			name: "complex action names with different versions",
			action1: models.Action{
				Uses:        "docker/build-push-action",
				UsesVersion: "v5.0.0",
			},
			action2: models.Action{
				Uses:        "docker/build-push-action",
				UsesVersion: "v5.1.0",
			},
			expected: true,
		},
		{
			name: "action with version tags (SHA vs semver)",
			action1: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v3",
			},
			action2: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "8e5e7e5ab8b370d6c329ec480221332ada57f0ab",
			},
			expected: true,
		},
		{
			name: "built-in steps (empty Uses) with different versions",
			action1: models.Action{
				Uses:        "",
				UsesVersion: "latest",
			},
			action2: models.Action{
				Uses:        "",
				UsesVersion: "v1",
			},
			expected: false,
		},
		{
			name: "action2 with empty Uses",
			action1: models.Action{
				Uses:        "actions/checkout",
				UsesVersion: "v3",
			},
			action2: models.Action{
				Uses:        "",
				UsesVersion: "v4",
			},
			expected: false,
		},
		{
			name: "version with different formats but same action",
			action1: models.Action{
				Uses:        "hashicorp/setup-terraform",
				UsesVersion: "v2",
			},
			action2: models.Action{
				Uses:        "hashicorp/setup-terraform",
				UsesVersion: "2.0.3",
			},
			expected: true,
		},
		{
			name: "latest vs specific version",
			action1: models.Action{
				Uses:        "actions/cache",
				UsesVersion: "latest",
			},
			action2: models.Action{
				Uses:        "actions/cache",
				UsesVersion: "v3",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasVersionDrift(tt.action1, tt.action2)
			if result != tt.expected {
				t.Errorf("HasVersionDrift(%+v, %+v) = %v, expected %v", tt.action1, tt.action2, result, tt.expected)
			}
		})
	}
}
