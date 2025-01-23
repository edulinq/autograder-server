package util

import (
	"strings"
)

// Count the lines of code in a file.
// Returns: (lines of code, total lines, error)
func LinesOfCode(path string) (int, int, error) {
	text, err := ReadFile(path)
	if err != nil {
		return 0, 0, err
	}

	// Note that Windows line endings are fine here, since they also have a newline.
	lines := strings.Split(text, "\n")

	loc := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			loc++
		}
	}

	return loc, len(lines), nil
}
