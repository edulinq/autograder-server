package util

import "log"

func GetSpecificEngineOptions(engineOptions map[string]any, engineName string) map[string]any {
	var specificEngineOptions map[string]any

	specificEngineOptionsAny, ok := engineOptions[engineName]
	if ok && specificEngineOptionsAny != nil {
		specificEngineOptions, ok = specificEngineOptionsAny.(map[string]any)
		if !ok {
			log.Printf("WARNING: Engine options for '%s' could not be converted into map[string]any. Engines will use default values.", engineName)
		}
	}
	return specificEngineOptions
}
