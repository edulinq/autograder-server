package model

import (
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/config"
)

func TestFullLoadCourseBase(test *testing.T) {
	courseID := "course101"
	testPath := filepath.Join(config.GetCourseImportDir(), "testdata", courseID, COURSE_CONFIG_FILENAME)

	course, submissions, err := FullLoadCourseFromPath(testPath)
	if err != nil {
		test.Fatalf("Failed to load course: '%v'.", err)
	}

	if courseID != course.ID {
		test.Fatalf("Unexpected course ID. Expected '%s', Actual: '%s'.", courseID, course.ID)
	}

	if len(course.Assignments) != 1 {
		test.Fatalf("Unexpected number of assignments. Expected %d, Actual: %d.", 1, len(course.Assignments))
	}

	if len(submissions) != 3 {
		test.Fatalf("Unexpected number of submissions. Expected %d, Actual: %d.", 3, len(submissions))
	}
}
