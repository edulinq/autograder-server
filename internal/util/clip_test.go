package util

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestClipString(test *testing.T) {
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
			name:     "Empty string",
			input:    "",
			maxLen:   100,
			expected: "",
		},
		{
			name:     "Exact length - no truncation",
			input:    strings.Repeat("b", 50),
			maxLen:   50,
			expected: strings.Repeat("b", 50),
		},
		{
			name:     "Long string - truncated",
			input:    strings.Repeat("a", 100),
			maxLen:   50,
			expected: "truncated",
		},
		{
			name:     "Too small maxLength returns empty",
			input:    "hello world",
			maxLen:   5,
			expected: "",
		},
		{
			name:     "Negative maxLength returns empty",
			input:    "hello world",
			maxLen:   -10,
			expected: "",
		},
		{
			name:     "Zero maxLength returns empty",
			input:    "hello world",
			maxLen:   0,
			expected: "",
		},
		{
			name:     "Unicode string - no corruption",
			input:    "Hello 世界! 🌍",
			maxLen:   20,
			expected: "Hello 世界! 🌍",
		},
		{
			name:     "Unicode string - truncated safely",
			input:    "Hello 世界! 🌍 " + strings.Repeat("x", 100),
			maxLen:   20,
			expected: "truncated",
		},
		{
			name:     "Just over limit",
			input:    strings.Repeat("a", 51),
			maxLen:   50,
			expected: "truncated",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			result := ClipString(testCase.input, testCase.maxLen)

			if testCase.expected == "truncated" {
				// Check that result contains truncation marker
				if !strings.Contains(result, "...") {
					test.Errorf("Test case '%s': Expected truncation marker, got: %q.", testCase.name, result)
				}
				// Check that result does not exceed maxLength in runes
				if utf8.RuneCountInString(result) > testCase.maxLen {
					test.Errorf("Test case '%s': Result length %d exceeds maxLen %d.", testCase.name, utf8.RuneCountInString(result), testCase.maxLen)
				}
				// Check original content is preserved at start
				if !strings.HasPrefix(result, "a") && testCase.input[0] == 'a' {
					test.Errorf("Test case '%s': Result should preserve start of original content.", testCase.name)
				}
			} else if result != testCase.expected {
				test.Errorf("Test case '%s': Expected: %q, Got: %q.", testCase.name, testCase.expected, result)
			}
		})
	}
}

func TestClipStringLengthInvariant(test *testing.T) {
	// Property-based test: result should never exceed maxLength
	testStrings := []string{
		"short",
		strings.Repeat("a", 1000),
		"Hello 世界! 🌍",
		"",
		strings.Repeat("🎉", 100),
	}

	maxLengths := []int{10, 50, 100, 500, 1000}

	for _, text := range testStrings {
		for _, maxLen := range maxLengths {
			result := ClipString(text, maxLen)
			if maxLen >= 10 { // minClipLength
				if utf8.RuneCountInString(result) > maxLen {
					test.Errorf("ClipString(%q, %d) = %q (len=%d), exceeds maxLength.",
						text, maxLen, result, utf8.RuneCountInString(result))
				}
			}
		}
	}
}

func TestClipStringWithHash(test *testing.T) {
	testCases := []struct {
		name   string
		input  string
		maxLen int
	}{
		{
			name:   "Short string - no truncation",
			input:  "hello world",
			maxLen: 100,
		},
		{
			name:   "Empty string",
			input:  "",
			maxLen: 100,
		},
		{
			name:   "Long string - truncated with hash",
			input:  strings.Repeat("a", 200),
			maxLen: 50,
		},
		{
			name:   "Too small maxLength falls back to simple clip",
			input:  "hello world",
			maxLen: 5,
		},
		{
			name:   "Unicode string",
			input:  "Hello 世界! 🌍 " + strings.Repeat("x", 100),
			maxLen: 50,
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			result := ClipStringWithHash(testCase.input, testCase.maxLen)

			if utf8.RuneCountInString(testCase.input) <= testCase.maxLen {
				if result != testCase.input {
					test.Errorf("Test case '%s': Expected: %q, Got: %q.", testCase.name, testCase.input, result)
				}
			} else if testCase.maxLen >= 10 {
				// Should be truncated
				if !strings.Contains(result, "...") {
					test.Errorf("Test case '%s': Expected truncation marker, got: %q.", testCase.name, result)
				}
				if utf8.RuneCountInString(result) > testCase.maxLen {
					test.Errorf("Test case '%s': Result length %d exceeds maxLen %d.", testCase.name, utf8.RuneCountInString(result), testCase.maxLen)
				}
			}
		})
	}
}

func TestClipStringWithHashDifferentHashes(test *testing.T) {
	// Test that different long strings produce different hashes
	str1 := strings.Repeat("a", 1000)
	str2 := strings.Repeat("b", 1000)
	maxLen := 50

	result1 := ClipStringWithHash(str1, maxLen)
	result2 := ClipStringWithHash(str2, maxLen)

	// Both should be truncated
	if !strings.Contains(result1, "...") || !strings.Contains(result2, "...") {
		test.Error("Both results should be truncated.")
	}

	// Hashes should be different
	if result1 == result2 {
		test.Error("Different strings should produce different clipped results.")
	}

	// Check lengths
	if utf8.RuneCountInString(result1) > maxLen || utf8.RuneCountInString(result2) > maxLen {
		test.Error("Results should not exceed maxLength.")
	}
}

func TestTruncateRunes(test *testing.T) {
	testCases := []struct {
		name     string
		input    string
		numRunes int
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			numRunes: 10,
			expected: "",
		},
		{
			name:     "Negative n",
			input:    "hello",
			numRunes: -1,
			expected: "",
		},
		{
			name:     "Zero n",
			input:    "hello",
			numRunes: 0,
			expected: "",
		},
		{
			name:     "n greater than length",
			input:    "hello",
			numRunes: 100,
			expected: "hello",
		},
		{
			name:     "ASCII truncation",
			input:    "hello world",
			numRunes: 5,
			expected: "hello",
		},
		{
			name:     "Unicode truncation - no corruption",
			input:    "Hello 世界",
			numRunes: 6,
			expected: "Hello ",
		},
		{
			name:     "Emoji truncation - no corruption",
			input:    "Hello 🌍🌍🌍",
			numRunes: 6,
			expected: "Hello ",
		},
		{
			name:     "Exact boundary",
			input:    "hello",
			numRunes: 5,
			expected: "hello",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			result := truncateRunes(testCase.input, testCase.numRunes)
			if result != testCase.expected {
				test.Errorf("Test case '%s': Expected: %q, Got: %q.", testCase.name, testCase.expected, result)
			}

			// Verify no UTF-8 corruption
			for _, r := range result {
				if r == utf8.RuneError {
					test.Errorf("Test case '%s': Result contains UTF-8 corruption (replacement character).", testCase.name)
					break
				}
			}
		})
	}
}
