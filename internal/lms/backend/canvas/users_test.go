package canvas

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestCanvasUserGetBase(test *testing.T) {
	testCases := []struct {
		email    string
		expected *lmstypes.User
	}{
		{
			"course-owner@test.edulinq.org",
			&lmstypes.User{
				ID:    "00010",
				Name:  "course-owner",
				Email: "course-owner@test.edulinq.org",
				Role:  model.CourseRoleOwner,
			},
		},
		{
			"course-admin@test.edulinq.org",
			&lmstypes.User{
				ID:    "00020",
				Name:  "course-admin",
				Email: "course-admin@test.edulinq.org",
				Role:  model.CourseRoleAdmin,
			},
		},
		{
			"course-student@test.edulinq.org",
			&lmstypes.User{
				ID:    "00040",
				Name:  "course-student",
				Email: "course-student@test.edulinq.org",
				Role:  model.CourseRoleStudent,
			},
		},
	}

	for i, testCase := range testCases {
		user, err := testBackend.FetchUser(testCase.email)
		if err != nil {
			test.Errorf("Case %d: Failed to fetch user: '%v'.", i, err)
			continue
		}

		if *testCase.expected != *user {
			test.Errorf("Case %d: User not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.expected, user)
			continue
		}
	}
}

func TestCanvasUsersGetBase(test *testing.T) {
	expected := []*lmstypes.User{
		&lmstypes.User{
			ID:    "00040",
			Name:  "course-student",
			Email: "course-student@test.edulinq.org",
			Role:  model.CourseRoleStudent,
		},
		&lmstypes.User{
			ID:    "00020",
			Name:  "course-admin",
			Email: "course-admin@test.edulinq.org",
			Role:  model.CourseRoleAdmin,
		},
		&lmstypes.User{
			ID:    "00010",
			Name:  "course-owner",
			Email: "course-owner@test.edulinq.org",
			Role:  model.CourseRoleOwner,
		},
	}

	users, err := testBackend.fetchUsers(true)
	if err != nil {
		test.Fatalf("Failed to fetch user: '%v'.", err)
	}

	if !reflect.DeepEqual(expected, users) {
		test.Fatalf("Users not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(users))
	}
}
