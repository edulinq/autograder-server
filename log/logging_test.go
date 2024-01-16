package log

import (
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
