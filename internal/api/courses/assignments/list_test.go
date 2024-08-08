package assignments

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestList(test *testing.T) {
	testCases := []struct {
		course string
	}{
		// Empty
		{"course-without-source"},
		// One Assignment
		{"course101-with-zero-limit"},
		// Multiple Assignments
		{"course-languages"},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"course-id": testCase.course,
		}

		response := core.SendTestAPIRequest(test, core.NewEndpoint(`courses/assignments/list`), fields)
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			continue
		}

		var responseContent ListResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		course := db.MustGetCourse(testCase.course)
		expectedInfos := core.NewAssignmentInfos(course.GetSortedAssignments())

		if !reflect.DeepEqual(expectedInfos, responseContent.Assignments) {
			test.Errorf("Case %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(expectedInfos), util.MustToJSONIndent(responseContent.Assignments))
			continue
		}
	}
}
