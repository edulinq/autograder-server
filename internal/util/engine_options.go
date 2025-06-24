package util

import (
	"github.com/edulinq/autograder/internal/log"
)

func ExtractEngineOptionMap(options any, engineName string, keys []string) (map[string]any, bool) {

	result := make(map[string]any) //declare the result variable

	if options == nil { //if options is nil, return
		return result, false
	}

	mapOptions, ok := options.(map[string]any) //check if  options is map[string]any
	if !ok {
		log.Warn("Engine options are not map[string]any.", log.NewAttr("engine", engineName))
		return result, false
	}

	engineSpecificAny, found := mapOptions[engineName] //check for engine in the map
	if !found {
		return result, false
	}

	engineOptions, ok := engineSpecificAny.(map[string]any) //check the structure of engine
	if !ok {
		log.Warn("Engine-specific options are not map[string]any.", log.NewAttr("engine", engineName))
		return result, false
	}

	for _, key := range keys { //loop through the keys
		val, exists := engineOptions[key]
		if !exists {
			continue
		}
		if fVal, ok := val.(float64); ok { //extract the float64 value
			result[key] = fVal
		} else {
			log.Warn("Engine option key has unsupported type.", log.NewAttr("engine", engineName), log.NewAttr("key", key), log.NewAttr("value", val))
		}
	}

	return result, len(result) > 0
}
