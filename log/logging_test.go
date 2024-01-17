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
