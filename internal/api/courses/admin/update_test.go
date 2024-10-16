package admin

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
)

// Change the course name and ensure it is back after an update.
func TestUpdateCourse(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testName := "ZZZ"

	// Change the course's name.
	course := db.MustGetTestCourse()
	course.Name = testName
	db.MustSaveCourse(course)

	// Verify name change.
	course = db.MustGetTestCourse()
	if course.Name != testName {
		test.Fatalf("Test name was not saved.")
	}

	// Update.
	response := core.SendTestAPIRequest(test, core.makeFullAPIPath(`courses/admin/update`), nil)
	if !response.Success {
		test.Errorf("Response is not a success when it should be: '%v'.", response)
	}

	// Verify name change is gone.
	course = db.MustGetTestCourse()
	if course.Name == testName {
		test.Fatalf("Name change survived update.")
	}
}
