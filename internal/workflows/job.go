package workflows

import (
	"fmt"
	"slices"
	"strings"

	"github.com/glaslos/tlsh"
)

type Action struct {
	Id string `json:"id"`
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
	hash *tlsh.TLSH
}

func (action *Action) GenerateHash() {
	action1Hash := tlsh.New()
	foundConfig := false

	settingsString := mapToString(action.Settings)
	if settingsString != "" {
		_, err := action1Hash.Write([]byte(settingsString))
		foundConfig = true
		if err != nil {
			return
		}
	}

	envString := mapToString(action.Env)
	if envString != "" {
		_, err := action1Hash.Write([]byte(envString))
		foundConfig = true
		if err != nil {
			return
		}
	}

	withString := mapToString(action.With)
	if withString != "" {
		_, err := action1Hash.Write([]byte(withString))
		foundConfig = true
		if err != nil {
			return
		}
	}

	if foundConfig {
		action1Hash.Sum(nil)
		action.hash = action1Hash
	}
}

func mapToString(m map[string]string) string {
	var sb strings.Builder

	// process keys in sorted order
	keys := []string{}
	for k, _ := range m {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s;", k, m[k]))
	}

	return sb.String()
}
