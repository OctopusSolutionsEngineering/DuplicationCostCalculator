package workflows

type Action struct {
	// Uses is the identifier of the GitHub Action. This does not include the version. For example: "actions/checkout".
	Uses string `json:"uses"`
	// UsesVersion is the version of the GitHub Action.
	UsesVersion string `json:"uses_version"`
	// Settings is a map of all other settings defined for the action excluding 'env' and 'with'.
	Settings map[string]string `json:"settings"`
	// Env is a map of environment variables set for the action.
	Env map[string]string `json:"env"`
	// With is a map of input parameters provided to the action.
	With map[string]string `json:"with"`
	// A locality sensitive hash of the action configuration.
	Hash string `json:"hash"`
}
