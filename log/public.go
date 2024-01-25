package log

import (
    "os"
)

func Trace(message string, args ...any) {
    logToLevel(LevelTrace, message, args...);
}

func Debug(message string, args ...any) {
    logToLevel(LevelDebug, message, args...);
}

func Info(message string, args ...any) {
    logToLevel(LevelInfo, message, args...);
}

func Warn(message string, args ...any) {
    logToLevel(LevelWarn, message, args...);
}

func Error(message string, args ...any) {
    logToLevel(LevelError, message, args...);
}

func Fatal(message string, args ...any) {
    FatalWithCode(1, message, args...);
}

func FatalWithCode(code int, message string, args ...any) {
    SetBackgroundLogging(false);

    logToLevel(LevelFatal, message, args...);
    os.Exit(code);
}

func SetLevel(level LogLevel) {
    SetLevels(level, level);
}

func SetLevelTrace() {
    SetLevel(LevelTrace);
}

func SetLevelDebug() {
    SetLevel(LevelDebug);
}

func SetLevelInfo() {
    SetLevel(LevelInfo);
}

func SetLevelWarn() {
    SetLevel(LevelWarn);
}

func SetLevelError() {
    SetLevel(LevelError);
}

func SetLevelFatal() {
    SetLevel(LevelFatal);
}
