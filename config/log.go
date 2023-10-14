package config;

// For the defaulted getters, the defualt will be returned on ANY error
// (even if the key exists, but is of the wrong type).

import (
    "os"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func InitLogging() {
    if (LOG_PRETTY.GetBool()) {
        log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr});
    } else {
        log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger();
    }

    var rawLogLevel = LOG_LEVEL.GetString();
    level, err := zerolog.ParseLevel(rawLogLevel);
    if (err != nil) {
        log.Fatal().Err(err).Str("level", rawLogLevel).Msg("Failed to parse the logging level.");
    }

    if (DEBUG.GetBool() && (level > zerolog.DebugLevel)) {
        level = zerolog.DebugLevel;
    }

    zerolog.SetGlobalLevel(level);
}

func GetLoggingLevel() zerolog.Level {
    return zerolog.GlobalLevel();
}

func SetLoggingLevel(level zerolog.Level) {
    LOG_LEVEL.Set(level.String());
    InitLogging();
}

func SetLogLevelTrace() {
    SetLoggingLevel(zerolog.TraceLevel);
}

func SetLogLevelDebug() {
    SetLoggingLevel(zerolog.DebugLevel);
}

func SetLogLevelInfo() {
    SetLoggingLevel(zerolog.InfoLevel);
}

func SetLogLevelWarn() {
    SetLoggingLevel(zerolog.WarnLevel);
}

func SetLogLevelError() {
    SetLoggingLevel(zerolog.ErrorLevel);
}

func SetLogLevelFatal() {
    SetLoggingLevel(zerolog.FatalLevel);
}
