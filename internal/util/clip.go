package util

import (
	"fmt"
	"unicode/utf8"
)

const clipSuffix = "..."

// ClipString truncates a string to maxLength runes and returns it.
// If includeHash is true and truncation occurs, a short hash suffix is appended.
func ClipString(text string, maxLength int, includeHash bool) string {
	if maxLength <= 0 {
		return ""
	}

	numRunes := utf8.RuneCountInString(text)
	if numRunes <= maxLength {
		return text
	}

	suffix := clipSuffix
	if includeHash {
		hash := Sha256HexFromString(text)[:8]
		suffix = fmt.Sprintf("%s [hash: %s]", clipSuffix, hash)
	}

	suffixRunes := utf8.RuneCountInString(suffix)
	if maxLength <= suffixRunes {
		return truncateRunes(suffix, maxLength)
	}

	keepRunes := maxLength - suffixRunes
	truncated := truncateRunes(text, keepRunes)

	return truncated + suffix
}

// truncateRunes safely truncates a string to a rune limit without breaking UTF-8.
func truncateRunes(text string, runeLimit int) string {
	if runeLimit <= 0 {
		return ""
	}

	if utf8.RuneCountInString(text) <= runeLimit {
		return text
	}

	count := 0
	for i, _ := range text {
		if count >= runeLimit {
			return text[:i]
		}
		count++
	}

	return text
}
