package util

import (
	"reflect"

	"path/filepath"
	"testing"
)

func TestMatchFilesBase(test *testing.T) {
	courseDir := filepath.Join(RootDirForTesting(), "testdata", "course-languages")

	paths := [2]string{
		filepath.Join(courseDir, "cpp"),
		filepath.Join(courseDir, "java"),
	}

	matches, unmatches, err := MatchFiles(paths)
	if err != nil {
		test.Fatalf("Failed to get matches: '%v'.", err)
	}

	expectedMatches := []string{
		"assignment.json",
		"grader.sh",
		"test-submissions/not-implemented/test-submission.json",
		"test-submissions/solution/test-submission.json",
	}

	expectedUnmatches := [][2]string{
		[2]string{"", "Assignment.java"},
		[2]string{"", "Grader.java"},
		[2]string{"assignment.cpp", ""},
		[2]string{"assignment.h", ""},
		[2]string{"config.json", ""},
		[2]string{"grader.cpp", ""},
		[2]string{"", "test-submissions/not-implemented/Assignment.java"},
		[2]string{"test-submissions/not-implemented/assignment.cpp", ""},
		[2]string{"", "test-submissions/solution/Assignment.java"},
		[2]string{"test-submissions/solution/assignment.cpp", ""},
	}

	if !reflect.DeepEqual(expectedMatches, matches) {
		test.Fatalf("Matches not as expected. Expected: '%s', Actual: '%s'.",
			MustToJSONIndent(expectedMatches), MustToJSONIndent(matches))
	}

	if !reflect.DeepEqual(expectedUnmatches, unmatches) {
		test.Fatalf("Unmatches not as expected. Expected: '%s', Actual: '%s'.",
			MustToJSONIndent(expectedUnmatches), MustToJSONIndent(unmatches))
	}
}

func TestMkTempDirCleanup(test *testing.T) {
	cleanupTempDir, err := MkDirTempFull("test-util-dir-", true)
	if err != nil {
		test.Fatalf("Failed to create cleanup temp dir: '%v'.", err)
	}

	noCleanupTempDir, err := MkDirTempFull("test-util-dir-", false)
	if err != nil {
		test.Fatalf("Failed to create no-cleanup temp dir: '%v'.", err)
	}
	defer RemoveDirent(noCleanupTempDir)

	err = RemoveRecordedTempDirs()
	if err != nil {
		test.Fatalf("Failed to remove recorded temp dirs: '%v'.", err)
	}

	if PathExists(cleanupTempDir) {
		test.Fatalf("Cleanup temp dir exists when it should have been cleaned up.")
	}

	if !PathExists(noCleanupTempDir) {
		test.Fatalf("No-Cleanup temp dir does not exist when it should not have been cleaned up.")
	}
}
