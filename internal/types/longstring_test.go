package types

import (
	"fmt"
	"strings"
	"testing"
)

func TestLongString(test *testing.T) {
	testCases := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "Short string - no truncation",
			input:    "hello world",
			maxLen:   100,
			expected: "hello world",
		},
		{
			name:     "Long string - truncated with hash",
			input:    strings.Repeat("a", 2000),
			maxLen:   100,
			expected: "truncated",
		},
		{
			name:     "Empty string",
			input:    "",
			maxLen:   100,
			expected: "",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			longString := LongString(testCase.input)
			result := longString.String()

			if testCase.expected == "truncated" {
				if !strings.Contains(result, "...") || !strings.Contains(result, "[hash:") {
					test.Errorf("Test case '%s': Expected truncation with hash, got: %q.", testCase.name, result)
				}
			} else if result != testCase.expected {
				test.Errorf("Test case '%s': Expected: %q, Got: %q.", testCase.name, testCase.expected, result)
			}
		})
	}
}

func TestLongStringWithFmt(test *testing.T) {
	// Test that fmt.Sprintf uses the String() method
	large := strings.Repeat("x", 5000)
	longString := LongString(large)
	result := fmt.Sprintf("%s", longString)

	if len(result) > MAX_LOG_STRING_LENGTH+50 {
		test.Errorf("String was not truncated by fmt.Sprintf.")
	}

	if !strings.Contains(result, "...") {
		test.Errorf("Expected truncation marker in output.")
	}

	if !strings.Contains(result, "[hash:") {
		test.Errorf("Expected hash in output.")
	}
}
