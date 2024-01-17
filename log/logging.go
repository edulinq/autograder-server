package log

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "strings"
    "sync"
    "time"
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
    KEY_COURSE = "course"
    KEY_ASSIGNMENT = "assignment"
    KEY_USER = "user"
)

type Attr struct {
    Name string
    Value any
}

type Loggable interface {
    LogValue() *Attr;
}

// TODO: Maybe make sync containers for both stderr and backend.

type LogRecord struct {
    // Core Attributes
    Level LogLevel `json:"level"`
    Message string `json:"message"`
    Timestamp string `json:"timestamp"`
    Error error `json:"error,omitempty"`

    // Context Attributes
    Course string `json:"course,omitempty"`
    Assignment string `json:"assignment,omitempty"`
    User string `json:"user,omitempty"`

    // Additional Attributes
    Attributes map[string]any `json:"attributes,omitempty"`
}

type storageBackend interface {
    LogDirect(record *LogRecord);
}

var textLevel LogLevel = LevelInfo;
var backendLevel LogLevel = LevelInfo;

var textWriter io.StringWriter = os.Stderr;
var backend storageBackend = nil;

var textLock sync.Mutex;
var backendLock sync.Mutex;

// Set the two logging levels.
// The textLevel controls the level of the logger outputting to the text writer.
// The backendLevel controls the level of the logger outputting to the backend.
func SetLevels(newTextLevel LogLevel, newBackendLevel LogLevel) {
    textLevel = newTextLevel;
    backendLevel = newBackendLevel;
}

func SetTextWriter(newTextWriter io.StringWriter) {
    textWriter = newTextWriter;
}

func SetStorageBackend(newBackend storageBackend) {
    backend = newBackend;
}

func LogDirect(record *LogRecord) {
    logText(record);
    logBackend(record);
}

func logBackend(record *LogRecord) {
    if ((backend == nil) || (record == nil)) {
        return;
    }

    if (record.Level < backendLevel) {
        return;
    }

    go func(record *LogRecord) {
        backendLock.Lock();
        defer backendLock.Unlock();

        backend.LogDirect(record);
    }(record);
}

func logText(record *LogRecord) {
    if ((textWriter == nil) || (record == nil)) {
        return;
    }

    if (record.Level < textLevel) {
        return;
    }

    builder := strings.Builder{};

    builder.WriteString(fmt.Sprintf("%s [%5s] %s", record.Timestamp, record.Level.String(), record.Message));

    if (record.Course != "") {
        record.Attributes[KEY_COURSE] = record.Course;
    }

    if (record.Assignment != "") {
        record.Attributes[KEY_ASSIGNMENT] = record.Assignment;
    }

    if (record.User != "") {
        record.Attributes[KEY_USER] = record.User;
    }

    if (len(record.Attributes) > 0) {
        value := "";

        bytes, err := json.Marshal(record.Attributes);
        if (err != nil) {
            value = fmt.Sprintf("%+v", record.Attributes);
            Error("JSON encoding error on logging attributes: '%+v'.", err);
        } else {
            value = string(bytes);
        }

        builder.WriteString(" | ");
        builder.WriteString(value);

        delete(record.Attributes, KEY_COURSE);
        delete(record.Attributes, KEY_ASSIGNMENT);
        delete(record.Attributes, KEY_USER);
    }

    if (record.Error != nil) {
        builder.WriteString(" | ");
        builder.WriteString(record.Error.Error());
    }

    builder.WriteString("\n");

    textLock.Lock();
    defer textLock.Unlock();

    textWriter.WriteString(builder.String());
}

func Log(level LogLevel, message string, course string, assignment string, user string, logError error, attributes map[string]any) {
    record := &LogRecord{
        Level: level,
        Message: message,
        Timestamp: time.Now().Format(time.RFC3339),
        Error: logError,

        Course: course,
        Assignment: assignment,
        User: user,

        Attributes: attributes,
    };

    LogDirect(record);
}

// Parse logging information from standard arguments.
// Arguments must be either:
// nil, error, Loggable, Attr, or *Attr.
// If there is an error parsing the attributes,
// the best effort attributes will be returned and an error will be retutned.
func parseArgs(args ...any) (string, string, string, error, map[string]any, error) {
    var course string;
    var assignment string;
    var user string;
    var logError error;
    var attributes map[string]any = make(map[string]any);
    var err error;

    for i, arg := range args {
        if (arg == nil) {
            continue;
        }

        var attr *Attr = nil;

        switch argValue := arg.(type) {
            case error:
                logError = argValue;
            case Loggable:
                attr = argValue.LogValue();
            case Attr:
                attr = &argValue;
            case *Attr:
                attr = argValue;
            default:
                err = errors.Join(err, fmt.Errorf("Logging argument %d is an unknown type '%T': '%v'.", i, argValue, argValue));
        }

        if (attr == nil) {
            continue;
        }

        switch attr.Name {
            case KEY_COURSE:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a course key, but non-string value '%T': '%v'.", i, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (course != "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate course key. Old value: '%s', New value: '%s'.", i, course, value));
                }

                course = value;
            case KEY_ASSIGNMENT:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a assignment key, but non-string value '%T': '%v'.", i, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (assignment != "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate assignment key. Old value: '%s', New value: '%s'.", i, assignment, value));
                }

                assignment = value;
            case KEY_USER:
                value, ok := attr.Value.(string);
                if (!ok) {
                    err = errors.Join(fmt.Errorf("Logging argument %d has a user key, but non-string value '%T': '%v'.", i, attr.Value, attr.Value));
                    value = fmt.Sprintf("%v", attr.Value);
                }

                if (user != "") {
                    err = errors.Join(fmt.Errorf("Logging argument %d is a duplicate user key. Old value: '%s', New value: '%s'.", i, user, value));
                }

                user = value;
            default:
                attributes[attr.Name] = attr.Value
        }
    }

    return course, assignment, user, logError, attributes, err;
}

func logToLevel(level LogLevel, message string, args ...any) {
    course, assignment, user, logError, attributes, err := parseArgs(args...);
    if (err != nil) {
        Error("Failed to parse logging arguments.", err);
    }

    Log(level, message, course, assignment, user, logError, attributes);
}

func (this LogLevel) String() string {
    switch this {
        case LevelTrace:
            return "TRACE";
        case LevelDebug:
            return "DEBUG";
        case LevelInfo:
            return "INFO";
        case LevelWarn:
            return "WARN";
        case LevelError:
            return "ERROR";
        case LevelFatal:
            return "FATAL";
        default:
            return fmt.Sprintf("%d", int32(this));
    }
}

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
