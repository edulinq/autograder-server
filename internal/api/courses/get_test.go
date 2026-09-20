package courses

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestGetBase(test *testing.T) {
	testCases := []struct {
		email   string
		target  string
		locator string
	}{
		// Base
		{"course-other", "course101", ""},
		{"course-student", "course101", ""},
		{"course-grader", "course101", ""},
		{"course-admin", "course101", ""},
		{"course-owner", "course101", ""},
		{"server-admin", "course101", ""},
		{"server-owner", "course101", ""},

		// Bad Perms
		{"server-user", "course101", "-040"},
		{"server-creator", "course101", "-040"},

		// Missing
		{"server-admin", "ZZZ", "-018"},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"course-id": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, `courses/get`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.locator != "" {
				if testCase.locator != response.Locator {
					test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
						i, testCase.locator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.locator != "" {
			test.Errorf("Case %d: Did not get an expected error: '%s'.", i, testCase.locator)
			continue
		}

		var responseContent GetResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expectedFound := true
		if expectedFound != responseContent.Found {
			test.Errorf("Case %d: Found course does not match. Expected: '%v', actual: '%v'.", i, expectedFound, responseContent.Found)
			continue
		}

		expectedCourse := core.NewCourseInfo(db.MustGetCourse(testCase.target))

		if !reflect.DeepEqual(expectedCourse, responseContent.Course) {
			test.Errorf("Case %d: Unexpected course result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(expectedCourse), util.MustToJSONIndent(responseContent.Course))
			continue
		}
	}
}

func TestGetInactiveCourse(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testStatuses := map[string]*model.CourseStatus{
		"server-admin@test.edulinq.org": {
			Active:  false,
			Source:  model.StatusSourceServer,
			Owner:   "server-admin@test.edulinq.org",
			SetTime: timestamp.Now(),
		},
		"course-admin@test.edulinq.org": {
			Active:  true,
			Source:  model.StatusSourceCourse,
			Owner:   "course-admin@test.edulinq.org",
			SetTime: timestamp.Now(),
		},
	}

	course := db.MustGetCourse("course101")
	course.Statuses = testStatuses
	db.MustSaveCourse(course)

	fields := map[string]any{
		"course-id": "course101",
	}

	response := core.SendTestAPIRequestFull(test, `courses/get`, fields, nil, "course-admin")
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent GetResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if responseContent.Course.Active {
		test.Fatalf("Expected course to be inactive, but was instead active")
	}

	actualStatuses := responseContent.Course.Statuses

	if !reflect.DeepEqual(testStatuses, actualStatuses) {
		test.Fatalf("Unexpected statuses. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(testStatuses), util.MustToJSONIndent(actualStatuses))
	}
}
