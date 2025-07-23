package assignments

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/report"
	"github.com/edulinq/autograder/internal/util"
)

func TestCourseReport(test *testing.T) {
	users := db.MustGetServerUsers()

	testCases := []struct {
		email     string
		course    string
		permError bool
		locator   string
	}{
		// Permission Too Low
		{"course-student@test.edulinq.org", "course101", true, "-020"},
		{"course-other@test.edulinq.org", "course101", true, "-020"},

		// Valid Permission
		{"course-grader@test.edulinq.org", "course101", false, ""},
		{"course-admin@test.edulinq.org", "course101", false, ""},
		{"server-admin@test.edulinq.org", "course101", false, ""},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"user-email": testCase.email,
			"user-pass":  util.Sha256HexFromString(*users[testCase.email].Name),
			"course-id":  testCase.course,
		}

		response := core.SendTestAPIRequest(test, `courses/assignments/report`, fields)
		if !response.Success {
			if testCase.permError {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.permError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		var responseContent CourseReportResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		course := db.MustGetCourse(testCase.course)
		expected, err := report.GetCourseScoringReport(course)
		if err != nil {
			test.Errorf("Error fetching course report: %s.", err)
		}

		// Normalize ignored fields.
		for _, question := range expected.Assignments[0].Questions {
			question.MinString = ""
			question.MaxString = ""
			question.StdDevString = ""
			question.MeanString = ""
			question.MedianString = ""
		}

		if !reflect.DeepEqual(expected, responseContent.CourseReport) {
			test.Errorf("Case %d: Returning incorrect course report. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent.CourseReport))
			continue
		}
	}
}
