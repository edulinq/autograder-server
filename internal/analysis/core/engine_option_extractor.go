package core

import (
	"fmt"
)

func GetSpecificEngineOptions(engineOptions map[string]any, engineName string) (map[string]any, error) {
	var specificEngineOptions map[string]any

	specificEngineOptionsAny, ok := engineOptions[engineName]
	if !ok || specificEngineOptionsAny == nil {
		return nil, nil
	}
	specificEngineOptions, ok = specificEngineOptionsAny.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected options for '%s' to be of type map[string]any, but got '%T'", engineName, specificEngineOptionsAny)
	}
	return specificEngineOptions, nil
}
