package core

func GetSpecificEngineOptions(engineOptions map[string]any, engineName string) (map[string]any, bool) {
	var specificEngineOptions map[string]any

	specificEngineOptionsAny, ok := engineOptions[engineName]
	if !ok || specificEngineOptionsAny == nil {
		return specificEngineOptions, false
	} else {
		return specificEngineOptions, true
	}
}
