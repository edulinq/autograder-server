// A simple logging infrastructure that allows us to log directly to stderr (textWriter)
// and a backend (presumably a database).

package log

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "strings"
    "sync"
    "time"
)

const (
    PRETTY_TIME_FORMAT = time.RFC3339
)

type Record struct {
    // Core Attributes
    Level LogLevel `json:"level"`
    Message string `json:"message"`
    UnixMicro int64 `json:"unix-time"`
    Error error `json:"error,omitempty"`

    // Context Attributes
    Course string `json:"course,omitempty"`
    Assignment string `json:"assignment,omitempty"`
    User string `json:"user,omitempty"`

    // Additional Attributes
    Attributes map[string]any `json:"attributes,omitempty"`
}

type StorageBackend interface {
    LogDirect(record *Record) error;
}

var textWriter io.StringWriter = os.Stderr;
var backend StorageBackend = nil;

var textLock sync.Mutex;
var backendLock sync.Mutex;

// Option to log serially, generally only for testing.
var backgroundBackendLogging bool = true;

func SetTextWriter(newTextWriter io.StringWriter) {
    textWriter = newTextWriter;
}

func SetStorageBackend(newBackend StorageBackend) {
    backend = newBackend;
}

// Set whether to log to the backend in the backgroun and return the old value.
// Generally should only be used for testing and fatal logs.
func SetBackgroundLogging(value bool) bool {
    backendLock.Lock();
    defer backendLock.Unlock();

    oldValue := backgroundBackendLogging;
    backgroundBackendLogging = value;

    return oldValue;
}

func LogDirectRecord(record *Record) {
    logText(record);
    logBackend(record);
}

func logBackend(record *Record) {
    if ((backend == nil) || (record == nil)) {
        return;
    }

    if (record.Level < backendLevel) {
        return;
    }

    backendLog := func(record *Record) {
        backendLock.Lock();
        defer backendLock.Unlock();

        err := backend.LogDirect(record);
        if (err != nil) {
            errRecord := &Record{
                Level: LevelError,
                Message: "Failed to log to storage backend.",
                UnixMicro: time.Now().UnixMicro(),
                Error: err,
            };
            logText(errRecord);
        }
    }

    if (backgroundBackendLogging) {
        go backendLog(record);
    } else {
        backendLog(record);
    }
}

func logText(record *Record) {
    if ((textWriter == nil) || (record == nil)) {
        return;
    }

    if (record.Level < textLevel) {
        return;
    }

    builder := strings.Builder{};

    timestamp := time.UnixMicro(record.UnixMicro).Format(PRETTY_TIME_FORMAT);
    builder.WriteString(fmt.Sprintf("%s [%5s] %s", timestamp, record.Level.String(), record.Message));

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
    record := &Record{
        Level: level,
        Message: message,
        UnixMicro: time.Now().UnixMicro(),
        Error: logError,

        Course: course,
        Assignment: assignment,
        User: user,

        Attributes: attributes,
    };

    LogDirectRecord(record);
}

func logToLevel(level LogLevel, message string, args ...any) {
    course, assignment, user, logError, attributes, err := parseArgs(args...);
    if (err != nil) {
        Error("Failed to parse logging arguments.", err);
    }

    Log(level, message, course, assignment, user, logError, attributes);
}
