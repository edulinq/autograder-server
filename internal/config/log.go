package config

import (
	"github.com/edulinq/autograder/internal/log"
)

func InitLoggingFromConfig() {
	textLevel, textErr := log.ParseLevel(LOG_TEXT_LEVEL.Get())
	backendLevel, backendErr := log.ParseLevel(LOG_BACKEND_LEVEL.Get())

	log.SetLevels(textLevel, backendLevel)

	if textErr != nil {
		log.Error("Failed to parse the logging level, setting to INFO.", textErr)
	}

	if backendErr != nil {
		log.Error("Failed to parse the logging level, setting to INFO.", backendErr)
	}
}
