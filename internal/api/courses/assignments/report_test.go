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
	testCases := []struct {
		email   string
		locator string
	}{
		// Permission Too Low
		{"course-student@test.edulinq.org", "-020"},
		{"course-other@test.edulinq.org", "-020"},

		// Valid Permission
		{"course-grader@test.edulinq.org", ""},
		{"course-admin@test.edulinq.org", ""},
		{"server-admin@test.edulinq.org", ""},
	}

	for i, testCase := range testCases {

		response := core.SendTestAPIRequestFull(test, `courses/assignments/report`, map[string]any{}, []string{}, testCase.email)
		if !response.Success {
			if len(testCase.locator) != 0 {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.locator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if len(testCase.locator) != 0 {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		var responseContent CourseReportResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		course := db.MustGetCourse("course101")
		expected, err := report.GetCourseScoringReport(course)
		if err != nil {
			test.Errorf("Error fetching course report: %s.", err)
			continue
		}

		// Normalize JSON ignored fields.
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
