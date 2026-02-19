package util

import (
	"fmt"
	"unicode/utf8"
)

// minClipLength is the minimum practical length for clipping.
// Below this, we just return an empty string.
const minClipLength = 10

// ClipString truncates a string to maxLength and returns it.
// If the string is longer than maxLength, it returns a truncated version
// with a suffix indicating how much was truncated.
// Returns empty string if maxLength is too small to fit any meaningful content.
func ClipString(text string, maxLength int) string {
	if maxLength < minClipLength {
		return ""
	}

	// Count runes, not bytes, to avoid UTF-8 corruption
	numRunes := utf8.RuneCountInString(text)
	if numRunes <= maxLength {
		return text
	}

	// Calculate suffix
	suffix := fmt.Sprintf("... [truncated %d more runes]", numRunes)
	suffixRunes := utf8.RuneCountInString(suffix)

	// If suffix itself is too long, use a simpler one
	if suffixRunes >= maxLength {
		suffix = "..."
		suffixRunes = 3
	}

	// Calculate how many runes we can keep
	keepRunes := maxLength - suffixRunes
	if keepRunes < 0 {
		keepRunes = 0
	}

	// Safely truncate by runes
	truncated := truncateRunes(text, keepRunes)

	return truncated + suffix
}

// ClipStringWithHash truncates a string to maxLength and includes a hash of the full content.
// This allows people to see when long values differ.
// Returns empty string if maxLength is too small.
func ClipStringWithHash(text string, maxLength int) string {
	if maxLength < minClipLength {
		return ""
	}

	numRunes := utf8.RuneCountInString(text)
	if numRunes <= maxLength {
		return text
	}

	// Calculate hash suffix
	hash := Sha256HexFromString(text)[:8]
	suffix := fmt.Sprintf("... [hash: %s]", hash)
	suffixRunes := utf8.RuneCountInString(suffix)

	// If suffix is too long, fall back to simple truncation
	if suffixRunes >= maxLength {
		return ClipString(text, maxLength)
	}

	// Calculate how many runes we can keep
	keepRunes := maxLength - suffixRunes
	if keepRunes < 0 {
		keepRunes = 0
	}

	truncated := truncateRunes(text, keepRunes)

	return truncated + suffix
}

// truncateRunes safely truncates a string to n runes without breaking UTF-8.
// Returns the original string if n >= number of runes in text.
func truncateRunes(text string, n int) string {
	if n <= 0 {
		return ""
	}

	// Fast path: count runes as we go
	count := 0
	for i, _ := range text {
		if count >= n {
			return text[:i]
		}
		count++
	}

	// String is shorter than n runes
	return text
}
