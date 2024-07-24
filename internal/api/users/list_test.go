package users

import (
	"reflect"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
)

func TestList(test *testing.T) {
	expectedUsers := []core.ServerUserInfo{
		{
			Email: "admin@test.com",
			Name:  "admin",
			Role:  model.GetServerUserRole("admin"),
			Courses: map[string]core.EnrollmentInfo{
				"course-languages":          {CourseID: "course-languages", CourseName: "Course Using Different Languages.", Role: model.GetCourseUserRole("admin")},
				"course-with-lms":           {CourseID: "course-with-lms", CourseName: "Course With LMS", Role: model.GetCourseUserRole("admin")},
				"course-without-source":     {CourseID: "course-without-source", CourseName: "Course Without Source", Role: model.GetCourseUserRole("admin")},
				"course101":                 {CourseID: "course101", CourseName: "Course 101", Role: model.GetCourseUserRole("admin")},
				"course101-with-zero-limit": {CourseID: "course101-with-zero-limit", CourseName: "Course 101 - With Zero Limit", Role: model.GetCourseUserRole("admin")},
			},
		},
		{
			Email: "grader@test.com",
			Name:  "grader",
			Role:  model.GetServerUserRole("creator"),
			Courses: map[string]core.EnrollmentInfo{
				"course-languages":          {CourseID: "course-languages", CourseName: "Course Using Different Languages.", Role: model.GetCourseUserRole("grader")},
				"course-with-lms":           {CourseID: "course-with-lms", CourseName: "Course With LMS", Role: model.GetCourseUserRole("grader")},
				"course-without-source":     {CourseID: "course-without-source", CourseName: "Course Without Source", Role: model.GetCourseUserRole("grader")},
				"course101":                 {CourseID: "course101", CourseName: "Course 101", Role: model.GetCourseUserRole("grader")},
				"course101-with-zero-limit": {CourseID: "course101-with-zero-limit", CourseName: "Course 101 - With Zero Limit", Role: model.GetCourseUserRole("grader")},
			},
		},
		{
			Email: "no-lms-id@test.com",
			Name:  "no-lms-id",
			Role:  model.GetServerUserRole("user"),
			Courses: map[string]core.EnrollmentInfo{
				"course-languages":      {CourseID: "course-languages", CourseName: "Course Using Different Languages.", Role: model.GetCourseUserRole("admin")},
				"course-with-lms":       {CourseID: "course-with-lms", CourseName: "Course With LMS", Role: model.GetCourseUserRole("admin")},
				"course-without-source": {CourseID: "course-without-source", CourseName: "Course Without Source", Role: model.GetCourseUserRole("admin")},
			},
		},
		{
			Email: "other@test.com",
			Name:  "other",
			Role:  model.GetServerUserRole("user"),
			Courses: map[string]core.EnrollmentInfo{
				"course-languages":          {CourseID: "course-languages", CourseName: "Course Using Different Languages.", Role: model.GetCourseUserRole("other")},
				"course-with-lms":           {CourseID: "course-with-lms", CourseName: "Course With LMS", Role: model.GetCourseUserRole("other")},
				"course-without-source":     {CourseID: "course-without-source", CourseName: "Course Without Source", Role: model.GetCourseUserRole("other")},
				"course101":                 {CourseID: "course101", CourseName: "Course 101", Role: model.GetCourseUserRole("other")},
				"course101-with-zero-limit": {CourseID: "course101-with-zero-limit", CourseName: "Course 101 - With Zero Limit", Role: model.GetCourseUserRole("other")},
			},
		},
		{
			Email: "owner@test.com",
			Name:  "owner",
			Role:  model.GetServerUserRole("owner"),
			Courses: map[string]core.EnrollmentInfo{
				"course-languages":          {CourseID: "course-languages", CourseName: "Course Using Different Languages.", Role: model.GetCourseUserRole("owner")},
				"course-with-lms":           {CourseID: "course-with-lms", CourseName: "Course With LMS", Role: model.GetCourseUserRole("owner")},
				"course-without-source":     {CourseID: "course-without-source", CourseName: "Course Without Source", Role: model.GetCourseUserRole("owner")},
				"course101":                 {CourseID: "course101", CourseName: "Course 101", Role: model.GetCourseUserRole("owner")},
				"course101-with-zero-limit": {CourseID: "course101-with-zero-limit", CourseName: "Course 101 - With Zero Limit", Role: model.GetCourseUserRole("owner")},
			},
		},
		{
			Email: "student@test.com",
			Name:  "student",
			Role:  model.GetServerUserRole("user"),
			Courses: map[string]core.EnrollmentInfo{
				"course-languages":          {CourseID: "course-languages", CourseName: "Course Using Different Languages.", Role: model.GetCourseUserRole("student")},
				"course-with-lms":           {CourseID: "course-with-lms", CourseName: "Course With LMS", Role: model.GetCourseUserRole("student")},
				"course-without-source":     {CourseID: "course-without-source", CourseName: "Course Without Source", Role: model.GetCourseUserRole("student")},
				"course101":                 {CourseID: "course101", CourseName: "Course 101", Role: model.GetCourseUserRole("student")},
				"course101-with-zero-limit": {CourseID: "course101-with-zero-limit", CourseName: "Course 101 - With Zero Limit", Role: model.GetCourseUserRole("student")},
			},
		},
		{
			Email:   "server-admin@test.com",
			Name:    "server-admin",
			Role:    model.GetServerUserRole("admin"),
			Courses: map[string]core.EnrollmentInfo{},
		},
		{
			Email:   "server-creator@test.com",
			Name:    "server-creator",
			Role:    model.GetServerUserRole("creator"),
			Courses: map[string]core.EnrollmentInfo{},
		},
		{
			Email:   "server-owner@test.com",
			Name:    "server-owner",
			Role:    model.GetServerUserRole("owner"),
			Courses: map[string]core.EnrollmentInfo{},
		},
		{
			Email:   "server-user@test.com",
			Name:    "server-user",
			Role:    model.GetServerUserRole("user"),
			Courses: map[string]core.EnrollmentInfo{},
		},
	}

	response := core.SendTestAPIRequest(test, core.NewEndpoint(`users/list`), nil)

	if !response.Success {
		test.Fatalf("Response is not a success: '%v'.", response)
	}

	responseContent := response.Content.(map[string]any)

	if responseContent["users"] == nil {
		test.Fatalf("Got a nil user list.")
	}

	rawUsers := responseContent["users"].([]any)
	actualUsers := make([]core.ServerUserInfo, 0, len(rawUsers))

	for _, rawUser := range rawUsers {
		actualUsers = append(actualUsers, *core.ServerUserInfoFromMap(rawUser.(map[string]any)))
	}

	slices.SortFunc(expectedUsers, core.CompareServerUserInfo)
	slices.SortFunc(actualUsers, core.CompareServerUserInfo)

	if !reflect.DeepEqual(expectedUsers, actualUsers) {
		test.Fatalf("Users not as expected. Expected: '%+v', actual: '%+v'.", expectedUsers, actualUsers)
	}
}
