package types

import (
	"fmt"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestLongString(test *testing.T) {
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
			name:  "Long string - truncated with hash",
			input: strings.Repeat("a", 2000),
			expected: strings.Repeat("a", MAX_LOG_STRING_LENGTH-20) +
				fmt.Sprintf("... [hash: %s]", util.Sha256HexFromString(strings.Repeat("a", 2000))[:8]),
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for i, testCase := range testCases {
		result := LongString(testCase.input).String()

		if result != testCase.expected {
			test.Errorf("Case %d (%s): Expected: %q, Got: %q.", i, testCase.name, testCase.expected, result)
		}
	}
}

// TestLongStringWithFmt verifies fmt.Sprintf uses String() for LongString.
func TestLongStringWithFmt(test *testing.T) {
	large := strings.Repeat("x", 5000)
	hash := util.Sha256HexFromString(large)[:8]
	expected := strings.Repeat("x", MAX_LOG_STRING_LENGTH-20) + fmt.Sprintf("... [hash: %s]", hash)
	result := fmt.Sprintf("%s", LongString(large))

	if result != expected {
		test.Errorf("Unexpected formatted value. Expected: %q, Got: %q.", expected, result)
	}
}
