package log

// Well-known exit codes (sysexits.h).
const (
	EXIT_CONFIG = 78
)

func Trace(message string, args ...any) {
	LogToLevel(LevelTrace, message, args...)
}

func Debug(message string, args ...any) {
	LogToLevel(LevelDebug, message, args...)
}

func Info(message string, args ...any) {
	LogToLevel(LevelInfo, message, args...)
}

func Warn(message string, args ...any) {
	LogToLevel(LevelWarn, message, args...)
}

func Error(message string, args ...any) {
	LogToLevel(LevelError, message, args...)
}

func Fatal(message string, args ...any) {
	FatalWithCode(1, message, args...)
}

func FatalExitConfig(message string, args ...any) {
	FatalWithCode(EXIT_CONFIG, message, args...)
}

func FatalWithCode(code int, message string, args ...any) {
	SetBackgroundLogging(false)

	LogToLevel(LevelFatal, message, args...)
	ExitForUtil(code)
}

func SetLevel(level LogLevel) {
	SetLevels(level, level)
}

func SetLevelTrace() {
	SetLevel(LevelTrace)
}

func SetLevelDebug() {
	SetLevel(LevelDebug)
}

func SetLevelInfo() {
	SetLevel(LevelInfo)
}

func SetLevelWarn() {
	SetLevel(LevelWarn)
}

func SetLevelError() {
	SetLevel(LevelError)
}

func SetLevelFatal() {
	SetLevel(LevelFatal)
}
