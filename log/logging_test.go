package log

import (
    "fmt"
    "reflect"
    "strings"
    "testing"
)

func TestLogTextBase(test *testing.T) {
    buffer := strings.Builder{};

    oldTextWriter := textWriter;
    SetTextWriter(&buffer);
    defer SetTextWriter(oldTextWriter);

    SetLevelTrace();
    defer SetLevelInfo();

    Trace("trace");
    Debug("debug");
    Info("info");
    Warn("warn");
    Error("error");

    expectedLines := []string{
        "[TRACE] trace",
        "[DEBUG] debug",
        "[ INFO] info",
        "[ WARN] warn",
        "[ERROR] error",
    };

    lines := strings.Split(strings.TrimSpace(buffer.String()), "\n");
    if (len(lines) != len(expectedLines)) {
        test.Fatalf("Number of lines does not match. Expected: %d, Actual: %d.", len(expectedLines), len(lines));
    }

    for i, expectedLine := range expectedLines {
        line := strings.TrimSpace(lines[i]);

        // Remove the timestamp.
        _, line, _ = strings.Cut(line, " ");

        if (expectedLine != line) {
            test.Errorf("Case %d: Line does not match. Expected: '%s', Actual: '%s'.", i, expectedLine, line);
        }
    }
}

type backendLogger struct {
    records []*LogRecord
}

func (this *backendLogger) LogDirect(record *LogRecord) {
    this.records = append(this.records, record);
}

// Test both passing records to the backend, and the content of the records (used for both loggers).
func TestBackendLogging(test *testing.T) {
    SetLevels(LevelOff, LevelTrace);
    defer SetLevelInfo();

    var backend backendLogger;
    SetStorageBackend(&backend);
    defer SetStorageBackend(nil);

    backgroundBackendLoggingForTesting = false;
    defer func() {
        backgroundBackendLoggingForTesting = true;
    }();

    // Empty.
    Info("");

    // Levels.
    Trace("trace");
    Debug("debug");
    Info("info");
    Warn("warn");
    Error("error");

    // Context.
    Info("msg", fakeCourse{"C"});
    Info("msg", Attr{KEY_COURSE, "C"});
    Info("msg", fakeAssignment{"A"});
    Info("msg", Attr{KEY_ASSIGNMENT, "A"});
    Info("msg", fakeUser{"U"});
    Info("msg", Attr{KEY_USER, "U"});
    Info("msg", fakeCourse{"C"}, fakeAssignment{"A"}, fakeUser{"U"});

    // Error
    Info("msg", fmt.Errorf("err"));

    // Attributes.
    Info("msg", Attr{"ABC", 123});

    expectedRecords := []*LogRecord{
        // Empty.
        &LogRecord{
            LevelInfo,
            "",
            "", nil,
            "", "", "",
            map[string]any{},
        },

        // Levels.
        &LogRecord{
            LevelTrace,
            "trace",
            "", nil,
            "", "", "",
            map[string]any{},
        },
        &LogRecord{
            LevelDebug,
            "debug",
            "", nil,
            "", "", "",
            map[string]any{},
        },
        &LogRecord{
            LevelInfo,
            "info",
            "", nil,
            "", "", "",
            map[string]any{},
        },
        &LogRecord{
            LevelWarn,
            "warn",
            "", nil,
            "", "", "",
            map[string]any{},
        },
        &LogRecord{
            LevelError,
            "error",
            "", nil,
            "", "", "",
            map[string]any{},
        },

        // Context.
        &LogRecord{
            LevelInfo,
            "msg",
            "", nil,
            "C", "", "",
            map[string]any{},
        },
        &LogRecord{
            LevelInfo,
            "msg",
            "", nil,
            "C", "", "",
            map[string]any{},
        },
        &LogRecord{
            LevelInfo,
            "msg",
            "", nil,
            "", "A", "",
            map[string]any{},
        },
        &LogRecord{
            LevelInfo,
            "msg",
            "", nil,
            "", "A", "",
            map[string]any{},
        },
        &LogRecord{
            LevelInfo,
            "msg",
            "", nil,
            "", "", "U",
            map[string]any{},
        },
        &LogRecord{
            LevelInfo,
            "msg",
            "", nil,
            "", "", "U",
            map[string]any{},
        },
        &LogRecord{
            LevelInfo,
            "msg",
            "", nil,
            "C", "A", "U",
            map[string]any{},
        },

        // Error.
        &LogRecord{
            LevelInfo,
            "msg",
            "", fmt.Errorf("err"),
            "", "", "",
            map[string]any{},
        },

        // Attributes.
        &LogRecord{
            LevelInfo,
            "msg",
            "", nil,
            "", "", "",
            map[string]any{"ABC": 123},
        },
    };

    if (len(backend.records) != len(expectedRecords)) {
        test.Fatalf("Number of records does not match. Expected: %d, Actual: %d.", len(expectedRecords), len(backend.records));
    }

    for i, expectedRecord := range expectedRecords {
        // Remove the timestamp.
        backend.records[i].Timestamp = "";

        if (!reflect.DeepEqual(expectedRecord, backend.records[i])) {
            test.Errorf("Case %d: Record does not match. Expected: '%+v', Actual: '%+v'.", i, expectedRecord, backend.records[i]);
        }
    }
}

type parseResults struct {
    course string
    assignment string
    user string
    logError error
    attributes map[string]any
}

func TestParseArgs(test *testing.T) {
    testCases := []struct{
        results parseResults
        err error
        args []any
    } {
        // Empty.
        {
            parseResults{"", "", "", nil, map[string]any{}},
            nil,
            []any{},
        },

        // All special args.
        {
            parseResults{"C", "", "", nil, map[string]any{}},
            nil,
            []any{fakeCourse{"C"}},
        },
        {
            parseResults{"", "A", "", nil, map[string]any{}},
            nil,
            []any{fakeAssignment{"A"}},
        },
        {
            parseResults{"", "", "U", nil, map[string]any{}},
            nil,
            []any{fakeUser{"U"}},
        },
        {
            parseResults{"", "", "", fmt.Errorf("err"), map[string]any{}},
            nil,
            []any{fmt.Errorf("err")},
        },
        {
            parseResults{"", "", "", nil, map[string]any{"ABC": 123}},
            nil,
            []any{Attr{"ABC", 123}},
        },

        // Special key order.
        {
            parseResults{"C", "A", "", nil, map[string]any{}},
            nil,
            []any{fakeCourse{"C"}, fakeAssignment{"A"}},
        },
        {
            parseResults{"C", "A", "", nil, map[string]any{}},
            nil,
            []any{fakeAssignment{"A"}, fakeCourse{"C"}},
        },

        // All special keys.
        {
            parseResults{"C", "A", "U", nil, map[string]any{}},
            nil,
            []any{fakeAssignment{"A"}, fakeUser{"U"}, fakeCourse{"C"}},
        },

        // Special keys with direct attributes.
        {
            parseResults{"C", "A", "U", nil, map[string]any{}},
            nil,
            []any{Attr{KEY_COURSE, "C"}, Attr{KEY_ASSIGNMENT, "A"}, Attr{KEY_USER, "U"}},
        },

        // Multiple attributes.
        {
            parseResults{"", "", "", nil, map[string]any{"ABC": 123, "xyz": "foo"}},
            nil,
            []any{Attr{"ABC", 123}, Attr{"xyz", "foo"}},
        },

        // All types of args.
        {
            parseResults{"C", "A", "U", fmt.Errorf("err"), map[string]any{"ABC": 123}},
            nil,
            []any{fakeAssignment{"A"}, fakeUser{"U"}, fakeCourse{"C"}, fmt.Errorf("err"), Attr{"ABC", 123}},
        },

        // Dup key errors.
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 1 is a duplicate course key. Old value: 'C1', New value: 'C2'."),
            []any{fakeCourse{"C1"}, fakeCourse{"C2"}},
        },
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 1 is a duplicate course key. Old value: 'C1', New value: 'C2'."),
            []any{fakeCourse{"C1"}, Attr{KEY_COURSE, "C2"}},
        },
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 1 is a duplicate assignment key. Old value: 'A1', New value: 'A2'."),
            []any{fakeAssignment{"A1"}, fakeAssignment{"A2"}},
        },
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 1 is a duplicate assignment key. Old value: 'A1', New value: 'A2'."),
            []any{fakeAssignment{"A1"}, Attr{KEY_ASSIGNMENT, "A2"}},
        },
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 1 is a duplicate user key. Old value: 'U1', New value: 'U2'."),
            []any{fakeUser{"U1"}, fakeUser{"U2"}},
        },
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 1 is a duplicate user key. Old value: 'U1', New value: 'U2'."),
            []any{fakeUser{"U1"}, Attr{KEY_USER, "U2"}},
        },

        // Bad special key values.
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 0 has a course key, but non-string value 'int': '1'."),
            []any{Attr{KEY_COURSE, 1}},
        },
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 0 has a assignment key, but non-string value '<nil>': '<nil>'."),
            []any{Attr{KEY_ASSIGNMENT, nil}},
        },
        {
            parseResults{"", "", "", nil, map[string]any{}},
            fmt.Errorf("Logging argument 0 has a user key, but non-string value '[]string': '[U]'."),
            []any{Attr{KEY_USER, []string{"U"}}},
        },
    }

    for i, testCase := range testCases {
        course, assignment, user, logError, attributes, err := parseArgs(testCase.args...);

        if (err != nil) {
            if (err.Error() != testCase.err.Error()) {
                test.Errorf("Case %d: Got an unexpected parse error. Expected: '%v', Actual: '%v'.", i, testCase.err, err);
            }

            continue;
        }

        results := parseResults{course, assignment, user, logError, attributes};
        if (!reflect.DeepEqual(testCase.results, results)) {
            test.Errorf("Case %d: Results are not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.results, results);
        }
    }
}

type fakeCourse struct {
    name string
}

func (this fakeCourse) LogValue() *Attr {
    return &Attr{KEY_COURSE, this.name};
}

type fakeAssignment struct {
    name string
}

func (this fakeAssignment) LogValue() *Attr {
    return &Attr{KEY_ASSIGNMENT, this.name};
}

type fakeUser struct {
    name string
}

func (this fakeUser) LogValue() *Attr {
    return &Attr{KEY_USER, this.name};
}
