package lmssync

import (
	// TEST
	"fmt"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	lmstest "github.com/edulinq/autograder/internal/lms/backend/test"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
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
		// TEST
		{true, true, true, 1, 4, 0, 1, 1},
	}

	for i, testCase := range testCases {
		reset()

		label := fmt.Sprintf("Attr: %v, Add: %v, Del: %v", testCase.syncAttributes, testCase.syncAdd, testCase.syncDel)

		lmstest.SetUsersModifier(testingUsers)
		course := db.MustGetTestCourse()

		course.GetLMSAdapter().SyncUserAttributes = testCase.syncAttributes
		course.GetLMSAdapter().SyncUserAdds = testCase.syncAdd
		course.GetLMSAdapter().SyncUserRemoves = testCase.syncDel

		/* TEST
		courseUsers, err := db.GetCourseUsers(course)
		if err != nil {
			test.Errorf("Case %d (%+v): Failed to get course users: '%v'.", i, testCase, err)
			continue
		}
		*/

		result, err := SyncAllLMSUsers(course, false, true)
		if err != nil {
			test.Errorf("Case %d (%s): User sync failed: '%v'.", i, label, err)
			continue
		}

		counts := model.GetUserOpResultsCounts(result)

		// TEST
		fmt.Println("###")
		fmt.Println(i)
		fmt.Println(label)
		fmt.Println(util.MustToJSONIndent(result))
		fmt.Println("---")
		fmt.Println(util.MustToJSONIndent(counts))
		fmt.Println("###")

		// Basic counts that are the same for all tests.

		expected := 6
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
			user.Role = model.RoleOwner
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
		Role:  model.RoleStudent,
	}
	users = append(users, addUser)

	return users
}

/* TEST
func TestCourseSyncLMSUsers(test *testing.T) {
	// Leave the db in a good state after the test.
	defer reset()

	for i, testCase := range getSyncLMSTestCases() {
		// Reload the test course every time.
		reset()

		lmstest.SetUsersModifier(testingUsers)
		course := db.MustGetTestCourse()

		course.GetLMSAdapter().SyncUserAttributes = testCase.syncAttributes
		course.GetLMSAdapter().SyncUserAdds = testCase.syncAdd
		course.GetLMSAdapter().SyncUserRemoves = testCase.syncDel

		courseUsers, err := db.GetCourseUsers(course)
		if err != nil {
			test.Errorf("Case %d (%+v): Failed to get course users: '%v'.", i, testCase, err)
			continue
		}

		email.ClearTestMessages()

		result, err := SyncAllLMSUsers(course, testCase.dryRun, testCase.sendEmails)
		if err != nil {
			test.Errorf("Case %d (%+v): User sync failed: '%v'.", i, testCase, err)
			continue
		}

		var unchangedUsers []*model.User = []*model.User{
			courseUsers["owner@test.com"],
		}

		testMessages := email.GetTestMessages()

		// LMS syncs cannot skip users.
		if len(result.Skip) != 0 {
			test.Errorf("Case %d (%+v): Skipped users is not empty.", i, testCase)
			continue
		}

		// There will always be mod users, since LMS IDs are always synced.
		// But when the option is on, additional attriutes will be synced.
		currentModUsers := modUsers
		if testCase.syncAttributes {
			currentModUsers = modAllUsers
		} else {
		}

		if !model.UsersPointerEqual(currentModUsers, result.Mod) {
			test.Errorf("Case %d (%+v): Unexpected mod users. Expected: '%s', actual: '%s'.",
				i, testCase, util.MustToJSON(currentModUsers), util.MustToJSON(result.Mod))
			continue
		}

		if testCase.syncAdd {
			if !model.UsersPointerEqual(addUsers, result.Add) {
				test.Errorf("Case %d (%+v): Unexpected add users. Expected: '%s', actual: '%s'.",
					i, testCase, util.MustToJSON(addUsers), util.MustToJSON(result.Add))
				continue
			}

			if len(result.Add) != len(result.ClearTextPasswords) {
				test.Errorf("Case %d (%+v): Number of cleartext passwords (%d) does not match number of add users (%d).",
					i, testCase, len(result.ClearTextPasswords), len(result.Add))
				continue
			}

			for _, user := range addUsers {
				_, ok := result.ClearTextPasswords[user.Email]
				if !ok {
					test.Errorf("Case %d (%+v): Add user '%s' does not have a cleartext password.", i, testCase, user.Email)
					continue
				}
			}

			if testCase.dryRun || !testCase.sendEmails {
				if len(testMessages) != 0 {
					test.Errorf("Case %d (%+v): User additions were enabled on a no-email/dry run, but %d new emails were found.", i, testCase, len(testMessages))
					continue
				}
			} else {
				if !email.ShallowSliceEqual(addEmails, testMessages) {
					test.Errorf("Case %d (%+v): Unexpected add emails. Expected: '%s', actual: '%s'.",
						i, testCase, util.MustToJSON(addEmails), util.MustToJSON(testMessages))
					continue
				}
			}
		} else {
			if len(result.Add) != 0 {
				test.Errorf("Case %d (%+v): User additions were disabled, but %d new users were found.", i, testCase, len(result.Add))
				continue
			}

			if len(result.ClearTextPasswords) != 0 {
				test.Errorf("Case %d (%+v): User additions were disabled, but %d new cleartext passwords were found.", i, testCase, len(result.ClearTextPasswords))
				continue
			}

			if len(testMessages) != 0 {
				test.Errorf("Case %d (%+v): User additions were disabled, but %d new emails were found.", i, testCase, len(testMessages))
				continue
			}
		}

		if testCase.syncDel {
			if !model.UsersPointerEqual(delUsers, result.Del) {
				test.Errorf("Case %d (%+v): Unexpected del users. Expected: '%s', actual: '%s'.",
					i, testCase, util.MustToJSON(delUsers), util.MustToJSON(result.Del))
				continue
			}
		} else {
			unchangedUsers = append(unchangedUsers, courseUsers["other@test.com"])

			if len(result.Del) != 0 {
				test.Errorf("Case %d (%+v): User deletions were disabled, but %d deleted users were found.", i, testCase, len(result.Del))
				continue
			}
		}

		if !model.UsersPointerEqual(unchangedUsers, result.Unchanged) {
			test.Errorf("Case %d (%+v): Unexpected unchanged users. Expected: '%s', actual: '%s'.",
				i, testCase, util.MustToJSON(unchangedUsers), util.MustToJSON(result.Unchanged))
			continue
		}
	}
}

// Get all possible test cases.
func getSyncLMSTestCases() []SyncLMSTestCase {
	return buildSyncLMSTestCase(nil, 0, make([]bool, 5))
}

func buildSyncLMSTestCase(testCases []SyncLMSTestCase, index int, currentCase []bool) []SyncLMSTestCase {
	if index >= len(currentCase) {
		return append(testCases, SyncLMSTestCase{
			dryRun:         currentCase[0],
			sendEmails:     currentCase[1],
			syncAttributes: currentCase[2],
			syncAdd:        currentCase[3],
			syncDel:        currentCase[4],
		})
	}

	currentCase[index] = true
	testCases = buildSyncLMSTestCase(testCases, index+1, currentCase)

	currentCase[index] = false
	testCases = buildSyncLMSTestCase(testCases, index+1, currentCase)

	return testCases
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
			user.Role = model.RoleOwner
		} else if user.Email == "owner@test.com" {
			// owner will not have anything changed (so we must manually remove their LMS ID).
			user.ID = ""
		}
	}

	users = slices.Delete(users, removeIndex, removeIndex+1)

	// Make an add user.
	addUser := &lmstypes.User{
		ID:    "lms-add@test.com",
		Name:  "add",
		Email: "add@test.com",
		Role:  model.RoleStudent,
	}
	users = append(users, addUser)

	return users
}

// The users that are marked as additions.
var addUsers []*model.User = []*model.User{
	&model.User{
		Email: "add@test.com",
		Name:  "add",
		Role:  model.RoleStudent,
		LMSID: "lms-add@test.com",
	},
}

// The users that are marked as deletions.
var delUsers []*model.User = []*model.User{
	&model.User{
		Email: "other@test.com",
		Name:  "other",
		Role:  model.RoleOther,
		LMSID: "",
	},
}

// All the users that are marked as mods.
var modAllUsers []*model.User = []*model.User{
	&model.User{
		Email: "student@test.com",
		Name:  "student",
		Role:  model.RoleStudent,
		LMSID: "lms-student@test.com",
	},
	&model.User{
		Email: "grader@test.com",
		Name:  "Changed Name",
		Role:  model.RoleGrader,
		LMSID: "lms-grader@test.com",
	},
	&model.User{
		Email: "admin@test.com",
		Name:  "admin",
		Role:  model.RoleOwner,
		LMSID: "lms-admin@test.com",
	},
}

// All the users that are marked as mods with no attribute syncing.
var modUsers []*model.User = []*model.User{
	&model.User{
		Email: "student@test.com",
		Name:  "student",
		Role:  model.RoleStudent,
		LMSID: "lms-student@test.com",
	},
	&model.User{
		Email: "grader@test.com",
		Name:  "grader",
		Role:  model.RoleGrader,
		LMSID: "lms-grader@test.com",
	},
	&model.User{
		Email: "admin@test.com",
		Name:  "admin",
		Role:  model.RoleAdmin,
		LMSID: "lms-admin@test.com",
	},
}

var addEmails []*email.Message = []*email.Message{
	&email.Message{
		To:      []string{"add@test.com"},
		Subject: "Autograder course101 -- User Account Created",
		HTML:    false,
	},
}
*/
