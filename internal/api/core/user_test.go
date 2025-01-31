package core

import (
	"reflect"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestNewServerUserInfos(test *testing.T) {
	users := db.MustGetServerUsers()

	testCases := []struct {
		serverUsers []*model.ServerUser
		expected    []*ServerUserInfo
	}{
		// ServerUserInfo, no course information.
		{
			[]*model.ServerUser{
				users["server-user@test.edulinq.org"],
				users["server-admin@test.edulinq.org"],
				users["server-owner@test.edulinq.org"],
			},
			[]*ServerUserInfo{
				{
					BaseUserInfo: BaseUserInfo{
						Type:  ServerUserInfoType,
						Email: "server-admin@test.edulinq.org",
						Name:  "server-admin",
					},
					Role:    model.GetServerUserRole("admin"),
					Courses: map[string]EnrollmentInfo{},
				},
				{
					BaseUserInfo: BaseUserInfo{
						Type:  ServerUserInfoType,
						Email: "server-owner@test.edulinq.org",
						Name:  "server-owner",
					},
					Role:    model.GetServerUserRole("owner"),
					Courses: map[string]EnrollmentInfo{},
				},
				{
					BaseUserInfo: BaseUserInfo{
						Type:  ServerUserInfoType,
						Email: "server-user@test.edulinq.org",
						Name:  "server-user",
					},
					Role:    model.GetServerUserRole("user"),
					Courses: map[string]EnrollmentInfo{},
				},
			},
		},

		// ServerUserInfos, course information.
		{
			[]*model.ServerUser{
				users["course-student@test.edulinq.org"],
				users["course-admin@test.edulinq.org"],
				users["course-grader@test.edulinq.org"],
			},
			[]*ServerUserInfo{
				{
					BaseUserInfo: BaseUserInfo{
						Type:  ServerUserInfoType,
						Email: "course-admin@test.edulinq.org",
						Name:  "course-admin",
					},
					Role: model.GetServerUserRole("user"),
					Courses: map[string]EnrollmentInfo{
						"course-languages": {
							CourseID: "course-languages",
							Role:     model.GetCourseUserRole("admin"),
						},
						"course101": {
							CourseID: "course101",
							Role:     model.GetCourseUserRole("admin"),
						},
					},
				},
				{
					BaseUserInfo: BaseUserInfo{
						Type:  ServerUserInfoType,
						Email: "course-grader@test.edulinq.org",
						Name:  "course-grader",
					},
					Role: model.GetServerUserRole("user"),
					Courses: map[string]EnrollmentInfo{
						"course-languages": {
							CourseID: "course-languages",
							Role:     model.GetCourseUserRole("grader"),
						},
						"course101": {
							CourseID: "course101",
							Role:     model.GetCourseUserRole("grader"),
						},
					},
				},
				{
					BaseUserInfo: BaseUserInfo{
						Type:  ServerUserInfoType,
						Email: "course-student@test.edulinq.org",
						Name:  "course-student",
					},
					Role: model.GetServerUserRole("user"),
					Courses: map[string]EnrollmentInfo{
						"course-languages": {
							CourseID: "course-languages",
							Role:     model.GetCourseUserRole("student"),
						},
						"course101": {
							CourseID: "course101",
							Role:     model.GetCourseUserRole("student"),
						},
					},
				},
			},
		},
	}

	for i, testCase := range testCases {
		serverUserInfos := make([]*ServerUserInfo, 0, len(testCase.serverUsers))
		for _, user := range testCase.serverUsers {
			serverUserInfos = append(serverUserInfos, NewServerUserInfo(user))
		}

		for _, serverUserInfo := range serverUserInfos {
			if serverUserInfo.Type != ServerUserInfoType {
				test.Errorf("Test %d: Unexpected server user info type. Expected: '%s', actual: '%s'.",
					i, ServerUserInfoType, serverUserInfo.Type)
				continue
			}
		}

		slices.SortFunc(serverUserInfos, CompareServerUserInfoPointer)

		if !reflect.DeepEqual(testCase.expected, serverUserInfos) {
			test.Errorf("Test %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(serverUserInfos))
			continue
		}
	}
}

func TestNewCourseUserInfos(test *testing.T) {
	users := db.MustGetServerUsers()

	testCases := []struct {
		serverUsers []*model.ServerUser
		expected    []*CourseUserInfo
	}{
		// CourseUserInfos, enrolled in course.
		{
			[]*model.ServerUser{
				users["course-student@test.edulinq.org"],
				users["course-admin@test.edulinq.org"],
				users["course-grader@test.edulinq.org"],
			},
			[]*CourseUserInfo{
				{
					BaseUserInfo: BaseUserInfo{
						Type:  CourseUserInfoType,
						Email: "course-admin@test.edulinq.org",
						Name:  "course-admin",
					},
					Role:  model.GetCourseUserRole("admin"),
					LMSID: "lms-course-admin@test.edulinq.org",
				},
				{
					BaseUserInfo: BaseUserInfo{
						Type:  CourseUserInfoType,
						Email: "course-grader@test.edulinq.org",
						Name:  "course-grader",
					},
					Role:  model.GetCourseUserRole("grader"),
					LMSID: "lms-course-grader@test.edulinq.org",
				},
				{
					BaseUserInfo: BaseUserInfo{
						Type:  CourseUserInfoType,
						Email: "course-student@test.edulinq.org",
						Name:  "course-student",
					},
					Role:  model.GetCourseUserRole("student"),
					LMSID: "lms-course-student@test.edulinq.org",
				},
			},
		},

		// CourseUserInfos, not enrolled, role escalation.
		{
			[]*model.ServerUser{
				users["server-admin@test.edulinq.org"],
				users["server-owner@test.edulinq.org"],
			},
			[]*CourseUserInfo{
				{
					BaseUserInfo: BaseUserInfo{
						Type:  CourseUserInfoType,
						Email: "server-admin@test.edulinq.org",
						Name:  "server-admin",
					},
					Role:  model.GetCourseUserRole("owner"),
					LMSID: "",
				},
				{
					BaseUserInfo: BaseUserInfo{
						Type:  CourseUserInfoType,
						Email: "server-owner@test.edulinq.org",
						Name:  "server-owner",
					},
					Role:  model.GetCourseUserRole("owner"),
					LMSID: "",
				},
			},
		},

		// CourseUserInfos, not enrolled, no role escalation.
		{
			[]*model.ServerUser{
				users["server-user@test.edulinq.org"],
				users["server-creator@test.edulinq.org"],
			},
			[]*CourseUserInfo{},
		},
	}

	for i, testCase := range testCases {
		courseUsers := make([]*model.CourseUser, 0, len(testCase.serverUsers))
		for _, user := range testCase.serverUsers {
			courseUser, err := user.ToCourseUser(db.TEST_COURSE_ID, true)
			if err != nil {
				test.Errorf("Test %d: Failed to convert a server user into a course user.", i)
				continue
			}

			if courseUser != nil {
				courseUsers = append(courseUsers, courseUser)
			}
		}

		courseUserInfos := make([]*CourseUserInfo, 0, len(courseUsers))
		for _, courseUser := range courseUsers {
			courseUserInfos = append(courseUserInfos, NewCourseUserInfo(courseUser))
		}

		for _, courseUserInfo := range courseUserInfos {
			if courseUserInfo.Type != CourseUserInfoType {
				test.Errorf("Test %d: Unexpected course user info type. Expected: '%s', actual: '%s'.",
					i, CourseUserInfoType, courseUserInfo.Type)
				continue
			}
		}

		slices.SortFunc(courseUserInfos, CompareCourseUserInfoPointer)

		if !reflect.DeepEqual(testCase.expected, courseUserInfos) {
			test.Errorf("Test %d: Unexpected result. Expected: '%s', actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(courseUserInfos))
			continue
		}
	}
}
