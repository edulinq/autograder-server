package admin

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

func TestUpdateCourse(test *testing.T) {
	// Remove a user and ensure the user is back after a reload.

	// Leave the course in a good state after the test.
	db.ResetForTesting()
	defer db.ResetForTesting()

	course := db.MustGetTestCourse()

	count := countAssignments(test, course)
	if count != 1 {
		test.Fatalf("Unexpected pre-remove assignment count. Expected 1, found %d.", count)
	}

	// Remove the assignment by deleting the key.
	delete(course.Assignments, db.TEST_ASSIGNMENT_ID)

	count = countAssignments(test, course)
	if count != 0 {
		test.Fatalf("Unexpected post-remove assignment count. Expected 0, found %d.", count)
	}

	reloadRequest(test)

	// Must get the course from the db to see updates.
	course = db.MustGetTestCourse()

	count = countAssignments(test, course)
	if count != 1 {
		test.Fatalf("Unexpected post-update assignment count. Expected 1, found %d.", count)
	}
}

func reloadRequest(test *testing.T) {
	response := core.SendTestAPIRequest(test, core.NewEndpoint(`courses/admin/update`), nil)
	if !response.Success {
		test.Errorf("Response is not a success when it should be: '%v'.", response)
	}
}

func countAssignments(test *testing.T, course *model.Course) int {
	assignments := course.GetAssignments()

	return len(assignments)
}
