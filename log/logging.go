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

    textLock.Lock();
    defer textLock.Unlock();

    textWriter.WriteString(record.String() + "\n");
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

func (this *Record) String() string {
    builder := strings.Builder{};

    timestamp := time.UnixMicro(this.UnixMicro).Format(PRETTY_TIME_FORMAT);
    builder.WriteString(fmt.Sprintf("%s [%5s] %s", timestamp, this.Level.String(), this.Message));

    attributes := make(map[string]any, len(this.Attributes) + 3);
    for key, value := range this.Attributes {
        attributes[key] = value;
    }

    if (this.Course != "") {
        attributes[KEY_COURSE] = this.Course;
    }

    if (this.Assignment != "") {
        attributes[KEY_ASSIGNMENT] = this.Assignment;
    }

    if (this.User != "") {
        attributes[KEY_USER] = this.User;
    }

    if (len(attributes) > 0) {
        value := "";

        bytes, err := json.Marshal(attributes);
        if (err != nil) {
            value = fmt.Sprintf("%+v", attributes);
            Error("JSON encoding error on logging attributes: '%+v'.", err);
        } else {
            value = string(bytes);
        }

        builder.WriteString(" | ");
        builder.WriteString(value);
    }

    if (this.Error != nil) {
        builder.WriteString(" | ");
        builder.WriteString(this.Error.Error());
    }

    return builder.String();
}
