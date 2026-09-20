package status

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestSet(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email            string
		force            bool
		existingStatuses map[string]*model.CourseStatus
		locator          string
		expectedSource   model.StatusSource
	}{
		// Basic Set
		{
			"course-admin",
			false,
			map[string]*model.CourseStatus{},
			"",
			model.StatusSourceCourse,
		},
		{
			"server-owner",
			false,
			map[string]*model.CourseStatus{},
			"",
			model.StatusSourceServer,
		},

		// Overwrite Without Force
		{
			"course-admin",
			false,
			map[string]*model.CourseStatus{
				"course-admin@test.edulinq.org": {
					Source: model.StatusSourceCourse,
				},
			},
			"-646",
			0,
		},

		// Overwrite With Force
		{
			"course-admin",
			true,
			map[string]*model.CourseStatus{
				"course-admin@test.edulinq.org": {
					Source: model.StatusSourceCourse,
				},
			},
			"",
			model.StatusSourceCourse,
		},

		// Invalid Permissions
		{
			"course-grader",
			false,
			map[string]*model.CourseStatus{},
			"-020",
			0,
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		course := db.MustGetCourse("course101")
		course.Statuses = testCase.existingStatuses
		db.MustSaveCourse(course)

		fields := map[string]any{
			"course-id": "course101",
			"active":    false,
			"force":     testCase.force,
		}

		response := core.SendTestAPIRequestFull(test, `courses/status/set`, fields, nil, testCase.email)
		if !response.Success {
			if testCase.locator != "" {
				if response.Locator != testCase.locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
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

		var responseContent SetResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if responseContent.Status.Source != testCase.expectedSource {
			test.Errorf("Case %d: Wrong source. Expected '%s', found '%s'.", i, testCase.expectedSource, responseContent.Status.Source)
		}
	}
}
