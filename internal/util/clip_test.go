package util

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestClipString(test *testing.T) {
	testCases := []struct {
		name        string
		input       string
		maxLength   int
		includeHash bool
		expected    string
	}{
		{
			name:        "No truncation",
			input:       "hello world",
			maxLength:   100,
			includeHash: false,
			expected:    "hello world",
		},
		{
			name:        "No truncation with hash enabled",
			input:       "hello world",
			maxLength:   100,
			includeHash: true,
			expected:    "hello world",
		},
		{
			name:        "Truncate plain ascii",
			input:       strings.Repeat("a", 20),
			maxLength:   10,
			includeHash: false,
			expected:    "aaaaaaa...",
		},
		{
			name:        "Truncate with hash suffix",
			input:       strings.Repeat("a", 40),
			maxLength:   25,
			includeHash: true,
			expected:    strings.Repeat("a", 5) + "... [hash: " + Sha256HexFromString(strings.Repeat("a", 40))[:8] + "]",
		},
		{
			name:        "Very small plain length",
			input:       strings.Repeat("a", 20),
			maxLength:   2,
			includeHash: false,
			expected:    "..",
		},
		{
			name:        "Unicode truncation",
			input:       "Hello 世界! 🌍",
			maxLength:   8,
			includeHash: false,
			expected:    "Hello...",
		},
		{
			name:        "Zero max length",
			input:       "hello world",
			maxLength:   0,
			includeHash: false,
			expected:    "",
		},
		{
			name:        "Negative max length",
			input:       "hello world",
			maxLength:   -10,
			includeHash: true,
			expected:    "",
		},
	}

	for i, testCase := range testCases {
		result := ClipString(testCase.input, testCase.maxLength, testCase.includeHash)

		if result != testCase.expected {
			test.Errorf("Case %d (%s): Expected: %q, Got: %q.", i, testCase.name, testCase.expected, result)
			continue
		}

		if (testCase.maxLength > 0) && (utf8.RuneCountInString(result) > testCase.maxLength) {
			test.Errorf("Case %d (%s): Result length %d exceeds maxLength %d.",
				i, testCase.name, utf8.RuneCountInString(result), testCase.maxLength)
		}

		for _, r := range result {
			if r == utf8.RuneError {
				test.Errorf("Case %d (%s): Result contains UTF-8 corruption.", i, testCase.name)
				break
			}
		}
	}
}
