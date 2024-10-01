package logs

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func TestQuery(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	oldValue := log.SetBackgroundLogging(false)
	defer log.SetBackgroundLogging(oldValue)

	log.SetLevels(log.LevelOff, log.LevelTrace)
	defer log.SetLevelFatal()

	// Don't bother for any in-depth testing
	// (that is done for each specific component).
	// Just make sure the API-level looks good.

	// Only looks for logs with this attribute.
	specialKey := "test-key"
	specialValue := "internal.api.logs.query.TestQuery"
	specialAttribute := log.NewAttr(specialKey, specialValue)

	log.Trace("trace", specialAttribute)
	log.Debug("debug", specialAttribute)
	log.Info("info", specialAttribute)
	log.Warn("warn", specialAttribute)
	log.Error("error", specialAttribute)

	courseID := "course101"
	log.Info("info", specialAttribute, db.MustGetCourse(courseID))

	assignmentID := "hw0"
	log.Info("info", specialAttribute, db.MustGetCourse(courseID), log.NewAssignmentAttr(assignmentID))

	email := "server-user@test.edulinq.org"
	log.Info("info", specialAttribute, db.MustGetServerUser(email))

	testCases := []struct {
		email                string
		query                log.RawLogQuery
		expectedCount        int
		permError            bool
		expectedErrorLocator string
	}{
		// Admin can do anything.
		{"server-admin@test.edulinq.org", log.RawLogQuery{}, 6, false, ""},
		{"server-admin@test.edulinq.org", log.RawLogQuery{LevelString: "trace"}, 8, false, ""},

		// Course admins can query their own course.
		{"course-admin@test.edulinq.org", log.RawLogQuery{CourseID: courseID}, 2, false, ""},
		{"course-admin@test.edulinq.org", log.RawLogQuery{}, 0, true, ""},

		// Grader's can't query courses.
		{"course-grader@test.edulinq.org", log.RawLogQuery{CourseID: courseID}, 0, true, ""},

		// Note that providing just an assignment is not enough, you need the course too.
		{"server-admin@test.edulinq.org", log.RawLogQuery{CourseID: courseID, AssignmentID: assignmentID}, 1, false, ""},
		{"server-admin@test.edulinq.org", log.RawLogQuery{AssignmentID: assignmentID}, 0, false, "-1100"},

		// User's can only query their own logs.
		{"server-user@test.edulinq.org", log.RawLogQuery{}, 0, true, ""},
		{"server-user@test.edulinq.org", log.RawLogQuery{TargetUser: email}, 1, false, ""},
	}

	for i, testCase := range testCases {
		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.query), &fields)

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`logs/query`), fields, nil, testCase.email)
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			continue
		}

		var responseContent QueryResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		// Only keep records with the special attribute.
		actualRecords := make([]*log.Record, 0)

		for _, record := range responseContent.Records {
			value, exists := record.Attributes[specialKey]
			if !exists {
				continue
			}

			if value != specialValue {
				test.Fatalf("Case %d: Found a log with a special key, but different special value. Expected: '%s', Actual: '%s'.", i, specialValue, value)
			}

			actualRecords = append(actualRecords, record)
		}

		if testCase.expectedCount != len(actualRecords) {
			test.Errorf("Case %d: Unexpected number of records. Expected: %d, Actual: %d.", i, testCase.expectedCount, len(actualRecords))
			continue
		}

		if (testCase.expectedErrorLocator == "") && (responseContent.Error == nil) {
			continue
		}

		if (testCase.expectedErrorLocator != "") && (responseContent.Error == nil) {
			test.Errorf("Case %d: Did not get expected error: '%s'.", i, testCase.expectedErrorLocator)
			continue
		}

		if (testCase.expectedErrorLocator == "") && (responseContent.Error != nil) {
			if testCase.permError {
				continue
			}

			test.Errorf("Case %d: Got unexpected error: '%s'.", i, util.MustToJSONIndent(responseContent.Error))
			continue
		}

		if testCase.permError {
			test.Errorf("Case %d: Perm error was not thrown when it was expected.", i)
			continue
		}

		if testCase.expectedErrorLocator != responseContent.Error.Locator {
			test.Errorf("Case %d: Error is not as expected. Expected: '%s', Actual: '%s'.", i, testCase.expectedErrorLocator, responseContent.Error.Locator)
			continue
		}
	}
}
