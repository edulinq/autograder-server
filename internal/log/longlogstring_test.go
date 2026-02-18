package log

import (
	"fmt"
	"strings"
	"testing"
)

func TestLongLogStringString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Short string - no truncation",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "Long string - truncated",
			input:    strings.Repeat("a", 2000),
			expected: strings.Repeat("a", 1000) + "... [truncated 1000 more bytes]",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Exact length - no truncation",
			input:    strings.Repeat("b", 1000),
			expected: strings.Repeat("b", 1000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lls := LongLogString(tc.input)
			result := lls.String()

			if result != tc.expected {
				t.Errorf("Expected length %d, got length %d", len(tc.expected), len(result))
				t.Errorf("Expected: %q", tc.expected)
				t.Errorf("Got: %q", result)
			}
		})
	}
}

func TestLongLogStringWithFmt(t *testing.T) {
	// Test that fmt.Sprintf uses the String() method
	large := strings.Repeat("x", 5000)
	lls := LongLogString(large)
	result := fmt.Sprintf("%s", lls)

	if len(result) > DEFAULT_MAX_LOG_STRING_LENGTH+50 {
		t.Errorf("String was not truncated by fmt.Sprintf")
	}

	if !strings.Contains(result, "... [truncated") {
		t.Errorf("Expected truncation marker in output")
	}
}
