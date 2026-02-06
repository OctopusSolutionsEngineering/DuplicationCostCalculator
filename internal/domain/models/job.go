package models

import (
	"github.com/OctopusSolutionsEngineering/DuplicationCostCalculator/internal/domain/collections"
	"github.com/glaslos/tlsh"
	"github.com/samber/lo"
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
	// This is used by script steps
	Run string `json:"run"`
	// A locality sensitive hash of the action configuration.
	Hash *tlsh.TLSH
}

func (action *Action) GenerateHash() {
	action1Hash := tlsh.New()
	foundConfig := false

	settingsString := collections.MapToString(action.Settings)
	if settingsString != "" {
		_, err := action1Hash.Write([]byte(settingsString))
		foundConfig = true
		if err != nil {
			return
		}
	}

	envString := collections.MapToString(action.Env)
	if envString != "" {
		_, err := action1Hash.Write([]byte(envString))
		foundConfig = true
		if err != nil {
			return
		}
	}

	withString := collections.MapToString(action.With)
	if withString != "" {
		_, err := action1Hash.Write([]byte(withString))
		foundConfig = true
		if err != nil {
			return
		}
	}

	if foundConfig {
		action1Hash.Sum(nil)

		if lo.EveryBy(action1Hash.Binary(), func(b byte) bool {
			return b == 0
		}) {
			return
		}

		// Small strings can not be hashed
		action.Hash = action1Hash
	}
}
