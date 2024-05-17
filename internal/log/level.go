package log

import (
    "fmt"
    "strings"
)

type LogLevel int32;

const (
    LevelTrace LogLevel = -20
    LevelDebug LogLevel = -10
    LevelInfo LogLevel = 0
    LevelWarn LogLevel = 10
    LevelError LogLevel = 20
    LevelFatal LogLevel = 30
    LevelOff LogLevel = 100
)

const (
    LEVEL_STRING_TRACE = "TRACE";
    LEVEL_STRING_DEBUG = "DEBUG";
    LEVEL_STRING_INFO = "INFO";
    LEVEL_STRING_WARN = "WARN";
    LEVEL_STRING_ERROR = "ERROR";
    LEVEL_STRING_FATAL = "FATAL";
    LEVEL_STRING_OFF = "OFF";
)

var textLevel LogLevel = LevelInfo;
var backendLevel LogLevel = LevelInfo;

// Set the two logging levels.
// The textLevel controls the level of the logger outputting to the text writer.
// The backendLevel controls the level of the logger outputting to the backend.
func SetLevels(newTextLevel LogLevel, newBackendLevel LogLevel) {
    textLevel = newTextLevel;
    backendLevel = newBackendLevel;
}

// Parse a logging level from text.
// Will return INFO and an erorr on error.
func ParseLevel(rawText string) (LogLevel, error) {
    text := strings.ToUpper(strings.TrimSpace(rawText));
    switch text {
        case LEVEL_STRING_TRACE:
            return LevelTrace, nil;
        case LEVEL_STRING_DEBUG:
            return LevelDebug, nil;
        case LEVEL_STRING_INFO:
            return LevelInfo, nil;
        case LEVEL_STRING_WARN:
            return LevelWarn, nil;
        case LEVEL_STRING_ERROR:
            return LevelError, nil;
        case LEVEL_STRING_FATAL:
            return LevelFatal, nil;
        case LEVEL_STRING_OFF:
            return LevelOff, nil;
        default:
            return LevelInfo, fmt.Errorf("Unknown log level '%s'.", rawText);
    };
}

func (this LogLevel) String() string {
    switch this {
        case LevelTrace:
            return LEVEL_STRING_TRACE;
        case LevelDebug:
            return LEVEL_STRING_DEBUG;
        case LevelInfo:
            return LEVEL_STRING_INFO;
        case LevelWarn:
            return LEVEL_STRING_WARN;
        case LevelError:
            return LEVEL_STRING_ERROR;
        case LevelFatal:
            return LEVEL_STRING_FATAL;
        case LevelOff:
            return LEVEL_STRING_OFF;
        default:
            return fmt.Sprintf("%d", int32(this));
    }
}
