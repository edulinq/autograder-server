package util

import (
	"fmt"
	"path/filepath"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"

	"github.com/edulinq/autograder/internal/log"
)

// Compute a diff (expected vs actual) in the context of our standard testing environment.
func MustComputeTestDiff(expected string, actual string) string {
	output, err := ComputeUnfiedDiffFull(expected, "expected", actual, "actual")
	if err != nil {
		log.Error("Failed to compute unified diff, falling back to full output.")
		output = fmt.Sprintf("--- expected ---\n%s\n----------------\n--- actual ---\n%s\n--------------", expected, actual)
	}

	return output
}

func ComputeUnfiedDiffFull(before string, beforeLabel string, after string, afterLabel string) (string, error) {
	tempDir, err := MkDirTemp("diff-")
	if err != nil {
		return "", fmt.Errorf("Failed to create temp dir for diff: '%w'.", err)
	}
	defer RemoveDirent(tempDir)

	path := filepath.Join(tempDir, "beforeLabel")
	edits := myers.ComputeEdits(span.URIFromPath(path), before, after)
	unified := gotextdiff.ToUnified(beforeLabel, afterLabel, before, edits)

	return fmt.Sprint(unified), nil
}
