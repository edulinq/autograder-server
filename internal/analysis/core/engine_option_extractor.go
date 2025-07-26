package core

import (
	"github.com/edulinq/autograder/internal/log"
)

func GetSpecificEngineOptions(engineOptions map[string]any, engineName string) (map[string]any, bool) {
	var specificEngineOptions map[string]any

	specificEngineOptionsAny, ok := engineOptions[engineName]
	if !ok || specificEngineOptionsAny == nil {
		return nil, false
	}
	specificEngineOptions, ok = specificEngineOptionsAny.(map[string]any)
	if !ok {
		log.Warn("Expected options for '%s' to be of type map[string]any, but got %T\n", engineName, specificEngineOptionsAny)
		return nil, false
	}

	return specificEngineOptions, true
}
