package status

import (
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestRemove(test *testing.T) {
	defer db.ResetForTesting()

	testCases := []struct {
		email            string
		target           string
		clear            bool
		existingStatuses map[string]*model.CourseStatus
		locator          string
		expectedOwners   []string
	}{
		// Self Deletion
		{
			"course-owner",
			"",
			false,
			map[string]*model.CourseStatus{
				"course-owner@test.edulinq.org": {Source: model.StatusSourceCourse},
			},
			"",
			[]string{
				"course-owner@test.edulinq.org",
			},
		},

		// Nothing to delete.
		{
			"course-owner",
			"",
			false,
			map[string]*model.CourseStatus{
				"course-admin@test.edulinq.org": {Source: model.StatusSourceCourse},
			},
			"",
			[]string{},
		},
		{
			"course-owner",
			"fake-user@test.edulinq.org",
			false,
			map[string]*model.CourseStatus{
				"course-admin@test.edulinq.org": {Source: model.StatusSourceCourse},
			},
			"",
			[]string{},
		},

		// Deleting and clearing targets with higher status.
		{
			"course-owner",
			"server-admin@test.edulinq.org",
			false,
			map[string]*model.CourseStatus{
				"server-admin@test.edulinq.org": {Source: model.StatusSourceServer},
			},
			"-648",
			nil,
		},
		{
			"course-admin",
			"",
			true,
			map[string]*model.CourseStatus{
				"course-admin@test.edulinq.org": {Source: model.StatusSourceCourse},
				"server-admin@test.edulinq.org": {Source: model.StatusSourceServer},
			},
			"",
			[]string{
				"course-admin@test.edulinq.org",
			},
		},

		// Clearing when the caller has the highest status.
		{
			"server-admin",
			"",
			true,
			map[string]*model.CourseStatus{
				"course-admin@test.edulinq.org": {Source: model.StatusSourceCourse},
				"server-admin@test.edulinq.org": {Source: model.StatusSourceServer},
				"server-owner@test.edulinq.org": {Source: model.StatusSourceServer},
			},
			"",
			[]string{
				"course-admin@test.edulinq.org",
				"server-admin@test.edulinq.org",
				"server-owner@test.edulinq.org",
			},
		},

		// Clearing when empty.
		{
			"server-owner",
			"",
			true,
			map[string]*model.CourseStatus{},
			"",
			[]string{},
		},

		// Invalid Permissions
		{
			"course-grader",
			"",
			false,
			map[string]*model.CourseStatus{},
			"-020",
			nil,
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		course := db.MustGetCourse("course101")
		course.Statuses = testCase.existingStatuses
		db.MustSaveCourse(course)

		fields := map[string]any{
			"course-id":    "course101",
			"target-owner": testCase.target,
			"clear":        testCase.clear,
		}

		response := core.SendTestAPIRequestFull(test, `courses/status/remove`, fields, nil, testCase.email)
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

		var responseContent RemoveResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		removedOwners := []string{}
		for owner := range responseContent.Removed {
			removedOwners = append(removedOwners, owner)
		}

		slices.Sort(removedOwners)

		if !slices.Equal(removedOwners, testCase.expectedOwners) {
			test.Errorf("Case %d: Expected '%v', found '%v'.", i, testCase.expectedOwners, removedOwners)
		}
	}
}
