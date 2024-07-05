package lmssync

import (
	"fmt"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	lmstest "github.com/edulinq/autograder/internal/lms/backend/test"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/model"
)

type SyncLMSTestCase struct {
	dryRun         bool
	sendEmails     bool
	syncAttributes bool
	syncAdd        bool
	syncDel        bool
}

func reset() {
	db.ResetForTesting()
	lmstest.ClearUsersModifier()
	email.ClearTestMessages()
}

func TestCourseSyncLMSUserEmails(test *testing.T) {
	reset()
	defer reset()

	course := db.MustGetTestCourse()

	course.GetLMSAdapter().SyncUserAttributes = true

	emails := []string{"student@test.com"}
	results, err := SyncLMSUserEmails(course, emails, false, false)
	if err != nil {
		test.Fatalf("Got an error when syncing known user: '%v'.", err)
	}

	if len(results) != 1 {
		test.Fatalf("Unexpected sync count. Expected 1, Actual: %d.", len(results))
	}

	result := results[0]

	if !result.Modified {
		test.Fatalf("Result modified flag not set.")
	}

	emails = []string{"ZZZ@test.com"}
	results, err = SyncLMSUserEmails(course, emails, false, false)
	if err != nil {
		test.Fatalf("Got an error when syncing unknown user: '%v'.", err)
	}

	if len(results) != 0 {
		test.Fatalf("Unexpected sync count. Expected 0, Actual: %d.", len(results))
	}
}

func TestCourseSyncLMSUsers(test *testing.T) {
	defer reset()

	testCases := []struct {
		syncAttributes bool
		syncAdd        bool
		syncDel        bool

		added    int
		modified int
		skipped  int
		enrolled int
		dropped  int
	}{
		{true, true, true, 1, 4, 0, 1, 1},
		{true, true, false, 1, 3, 1, 1, 0},
		{true, false, true, 0, 4, 1, 0, 1},
		{true, false, false, 0, 3, 2, 0, 0},
		{false, true, true, 1, 1, 4, 1, 1},
		{false, true, false, 1, 0, 5, 1, 0},
		{false, false, true, 0, 1, 5, 0, 1},
		{false, false, false, 0, 0, 5, 0, 0},
	}

	for i, testCase := range testCases {
		reset()

		label := fmt.Sprintf("Attr: %v, Add: %v, Del: %v", testCase.syncAttributes, testCase.syncAdd, testCase.syncDel)

		lmstest.SetUsersModifier(testingUsers)
		course := db.MustGetTestCourse()

		course.GetLMSAdapter().SyncUserAttributes = testCase.syncAttributes
		course.GetLMSAdapter().SyncUserAdds = testCase.syncAdd
		course.GetLMSAdapter().SyncUserRemoves = testCase.syncDel

		result, err := SyncAllLMSUsers(course, false, true)
		if err != nil {
			test.Errorf("Case %d (%s): User sync failed: '%v'.", i, label, err)
			continue
		}

		counts := model.GetUserOpResultsCounts(result)

		// Basic counts that are the same for all tests.

		expected := 6
		if !course.GetLMSAdapter().SyncUsers() {
			// When stopping early, deletes are not considered.
			expected = 5
		}

		if expected != counts.Total {
			test.Errorf("Case %d (%s): Unexpected total number of results. Expected: %d, Actual: %d.", i, label, expected, counts.Total)
			continue
		}

		expected = 0
		if expected != counts.Removed {
			test.Errorf("Case %d (%s): Unexpected number of removes. Expected: %d, Actual: %d.", i, label, expected, counts.Removed)
			continue
		}

		expected = 0
		if expected != counts.NotExists {
			test.Errorf("Case %d (%s): Unexpected number of not exists. Expected: %d, Actual: %d.", i, label, expected, counts.NotExists)
			continue
		}

		expected = 0
		if expected != counts.ValidationErrors {
			test.Errorf("Case %d (%s): Unexpected number of validation errors. Expected: %d, Actual: %d.", i, label, expected, counts.ValidationErrors)
			continue
		}

		expected = 0
		if expected != counts.SystemErrors {
			test.Errorf("Case %d (%s): Unexpected number of system errors. Expected: %d, Actual: %d.", i, label, expected, counts.SystemErrors)
			continue
		}

		// The number of adds, emailed, and cleartext passwords should always match for these cases.

		if counts.Added != counts.Emailed {
			test.Errorf("Case %d (%s): Add and email count do not match. Add: %d, Email: %d.", i, label, counts.Added, counts.Emailed)
			continue
		}

		if counts.Added != counts.CleartextPassword {
			test.Errorf("Case %d (%s): Add and cleartext password count do not match. Add: %d, Cleartext Passowrd: %d.", i, label, counts.Added, counts.CleartextPassword)
			continue
		}

		if counts.Emailed != len(email.GetTestMessages()) {
			test.Errorf("Case %d (%s): Email count and number of test messages do not match. Email: %d, Test Messages: %d.", i, label, counts.Emailed, len(email.GetTestMessages()))
			continue
		}

		// Non-computed variable counts.

		if testCase.added != counts.Added {
			test.Errorf("Case %d (%s): Unexpected number of added. Expected: %d, Actual: %d.", i, label, testCase.added, counts.Added)
			continue
		}

		if testCase.modified != counts.Modified {
			test.Errorf("Case %d (%s): Unexpected number of modified. Expected: %d, Actual: %d.", i, label, testCase.modified, counts.Modified)
			continue
		}

		if testCase.skipped != counts.Skipped {
			test.Errorf("Case %d (%s): Unexpected number of skipped. Expected: %d, Actual: %d.", i, label, testCase.skipped, counts.Skipped)
			continue
		}

		if testCase.enrolled != counts.Enrolled {
			test.Errorf("Case %d (%s): Unexpected number of enrolled. Expected: %d, Actual: %d.", i, label, testCase.enrolled, counts.Enrolled)
			continue
		}

		if testCase.dropped != counts.Dropped {
			test.Errorf("Case %d (%s): Unexpected number of dropped. Expected: %d, Actual: %d.", i, label, testCase.dropped, counts.Dropped)
			continue
		}
	}
}

// Modify the users that the LMS will return for testing.
func testingUsers(users []*lmstypes.User) []*lmstypes.User {
	// Remove other.
	removeIndex := -1
	for i, user := range users {
		if user.Email == "other@test.com" {
			removeIndex = i
		} else if user.Email == "student@test.com" {
			// student will only have their LMS ID added, no other changes.
		} else if user.Email == "grader@test.com" {
			// grader will have their name changed.
			user.Name = "Changed Name"
		} else if user.Email == "admin@test.com" {
			// admin will have their role changed.
			user.Role = model.CourseRoleOwner
		} else if user.Email == "owner@test.com" {
			// owner will not have anything changed (so we must manually remove their email).
			user.ID = ""
		}
	}

	users = slices.Delete(users, removeIndex, removeIndex+1)

	// Make an add user.
	addUser := &lmstypes.User{
		ID:    "lms-add@test.com",
		Name:  "add",
		Email: "add@test.com",
		Role:  model.CourseRoleStudent,
	}
	users = append(users, addUser)

	return users
}
