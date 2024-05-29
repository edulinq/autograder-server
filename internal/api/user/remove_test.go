package user

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestUserRemove(test *testing.T) {
	// Leave the course in a good state after the test.
	defer db.ResetForTesting()

	testCases := []struct {
		role           model.CourseUserRole
		target         string
		basicPermError bool
		advPermError   bool
		expected       RemoveResponse
	}{
		{model.RoleOwner, "other@test.com", false, false, RemoveResponse{true}},
		{model.RoleOwner, "student@test.com", false, false, RemoveResponse{true}},
		{model.RoleOwner, "grader@test.com", false, false, RemoveResponse{true}},
		{model.RoleOwner, "admin@test.com", false, false, RemoveResponse{true}},
		{model.RoleOwner, "owner@test.com", false, false, RemoveResponse{true}},

		{model.RoleOther, "other@test.com", true, false, RemoveResponse{true}},
		{model.RoleStudent, "other@test.com", true, false, RemoveResponse{true}},
		{model.RoleGrader, "other@test.com", true, false, RemoveResponse{true}},
		{model.RoleAdmin, "other@test.com", false, false, RemoveResponse{true}},
		{model.RoleOwner, "other@test.com", false, false, RemoveResponse{true}},

		{model.RoleOwner, "ZZZ", false, false, RemoveResponse{false}},

		{model.RoleAdmin, "owner@test.com", false, true, RemoveResponse{true}},
		{model.RoleOwner, "owner@test.com", false, false, RemoveResponse{true}},
	}

	for i, testCase := range testCases {
		// Reload the test course every time.
		db.ResetForTesting()

		fields := map[string]any{
			"target-email": testCase.target,
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`user/remove`), fields, nil, testCase.role)

		if !response.Success {
			expectedLocator := ""
			if testCase.basicPermError {
				expectedLocator = "-020"
			} else if testCase.advPermError {
				expectedLocator = "-801"
			}

			if expectedLocator == "" {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			} else {
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			}

			continue
		}

		var responseContent RemoveResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if testCase.expected != responseContent {
			test.Errorf("Case %d: Unexpected result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent)
			continue
		}

		if !testCase.expected.FoundUser {
			continue
		}

		// Ensure that the user is removed.

		course := db.MustGetCourse("course101")
		user, err := db.GetUser(course, testCase.target)
		if err != nil {
			test.Errorf("Case %d: Failed to get removed user: '%v'.", i, err)
			continue
		}

		if user != nil {
			test.Errorf("Case %d: User not removed from DB: '%+v'.", i, user)
			continue
		}
	}
}
