package db

import (
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/model"
)

// Update a course from a path source.
func (this *DBTests) DBTestCourseUpdateCourseFromSourceBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()

	count := countAssignments(test, course)
	if count != 1 {
		test.Fatalf("Unexpected pre-remove assignment count. Expected 1, found %d.", count)
	}

	// Remove the assignment by deleting the key.
	delete(course.Assignments, TEST_ASSIGNMENT_ID)

	count = countAssignments(test, course)
	if count != 0 {
		test.Fatalf("Unexpected post-remove assignment count. Expected 0, found %d.", count)
	}

	newCourse, updated, err := UpdateCourseFromSource(course)
	if err != nil {
		test.Fatalf("Failed to update course: '%v'.", err)
	}

	if !updated {
		test.Fatalf("Course did not update.")
	}

	count = countAssignments(test, newCourse)
	if count != 1 {
		test.Fatalf("Unexpected post-update assignment count. Expected 1, found %d.", count)
	}
}

// Set the course's source to nil and then update.
// This will cause the course to skip updating.
func (this *DBTests) DBTestCourseUpdateCourseFromSourceSkip(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()

	count := countAssignments(test, course)
	if count != 1 {
		test.Fatalf("Unexpected pre-remove assignment count. Expected 1, found %d.", count)
	}

	// Remove the assignment by deleting the key.
	delete(course.Assignments, TEST_ASSIGNMENT_ID)

	count = countAssignments(test, course)
	if count != 0 {
		test.Fatalf("Unexpected post-remove assignment count. Expected 0, found %d.", count)
	}

	// Set the source to nil.
	course.Source = common.GetNilFileSpec()
	err := SaveCourse(course)
	if err != nil {
		test.Fatalf("Failed to save course: '%v'.", err)
	}

	_, updated, err := UpdateCourseFromSource(course)
	if err != nil {
		test.Fatalf("Failed to update course: '%v'.", err)
	}

	if updated {
		test.Fatalf("Course was (incorrectly) updated.")
	}

	// We can actually use the old course to still get a count.
	count = countAssignments(test, course)
	if count != 0 {
		test.Fatalf("Unexpected post-update assignment count. Expected 0, found %d.", count)
	}
}

func countAssignments(test *testing.T, course *model.Course) int {
	assignments := course.GetAssignments()

	return len(assignments)
}
