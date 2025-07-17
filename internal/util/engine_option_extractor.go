package util

import "log"

func GetSpecificEngineOptions(engineOptions map[string]any, engineName string) (map[string]any, bool) {
	var specificEngineOptions map[string]any

	specificEngineOptionsAny, ok := engineOptions[engineName]
	if ok && specificEngineOptionsAny != nil {
		specificEngineOptions, ok = specificEngineOptionsAny.(map[string]any)
		if !ok {
			log.Printf("Engine options for '%s' could not be converted into map[string]any. Engines will use default values.", engineName)
			return specificEngineOptions, false
		}
	} else {
		log.Printf("No engine options found for '%s'. Engines will use default values.", engineName)
		return specificEngineOptions, false
	}
	return specificEngineOptions, true
}
