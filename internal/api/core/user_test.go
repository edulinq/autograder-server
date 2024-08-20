package core

import (
	"reflect"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestNewUserInfos(test *testing.T) {
	users := db.MustGetServerUsers()

	serverUsers := []*model.ServerUser{
		users["student@test.com"],
		users["admin@test.com"],
		users["grader@test.com"],
	}

	testCases := []struct {
		toCourseUser   bool
		serverExpected []*ServerUserInfo
		courseExpected []*CourseUserInfo
	}{
		// Compare ServerUserInfos.
		{
			false,
			[]*ServerUserInfo{
				{
					UserInfo: UserInfo{
						Type:  ServerUserInfoType,
						Email: "admin@test.com",
						Name:  "admin",
					},
					Role: model.GetServerUserRole("user"),
					Courses: map[string]EnrollmentInfo{
						"course-languages": {
							CourseID:   "course-languages",
							CourseName: "Course Using Different Languages.",
							Role:       model.GetCourseUserRole("admin"),
						},
						"course-with-lms": {
							CourseID:   "course-with-lms",
							CourseName: "Course With LMS",
							Role:       model.GetCourseUserRole("admin"),
						},
						"course-without-source": {
							CourseID:   "course-without-source",
							CourseName: "Course Without Source",
							Role:       model.GetCourseUserRole("admin"),
						},
						"course101": {
							CourseID:   "course101",
							CourseName: "Course 101",
							Role:       model.GetCourseUserRole("admin"),
						},
						"course101-with-zero-limit": {
							CourseID:   "course101-with-zero-limit",
							CourseName: "Course 101 - With Zero Limit",
							Role:       model.GetCourseUserRole("admin"),
						},
					},
				},
				{
					UserInfo: UserInfo{
						Type:  ServerUserInfoType,
						Email: "grader@test.com",
						Name:  "grader",
					},
					Role: model.GetServerUserRole("user"),
					Courses: map[string]EnrollmentInfo{
						"course-languages": {
							CourseID:   "course-languages",
							CourseName: "Course Using Different Languages.",
							Role:       model.GetCourseUserRole("grader"),
						},
						"course-with-lms": {
							CourseID:   "course-with-lms",
							CourseName: "Course With LMS",
							Role:       model.GetCourseUserRole("grader"),
						},
						"course-without-source": {
							CourseID:   "course-without-source",
							CourseName: "Course Without Source",
							Role:       model.GetCourseUserRole("grader"),
						},
						"course101": {
							CourseID:   "course101",
							CourseName: "Course 101",
							Role:       model.GetCourseUserRole("grader"),
						},
						"course101-with-zero-limit": {
							CourseID:   "course101-with-zero-limit",
							CourseName: "Course 101 - With Zero Limit",
							Role:       model.GetCourseUserRole("grader"),
						},
					},
				},
				{
					UserInfo: UserInfo{
						Type:  ServerUserInfoType,
						Email: "student@test.com",
						Name:  "student",
					},
					Role: model.GetServerUserRole("user"),
					Courses: map[string]EnrollmentInfo{
						"course-languages": {
							CourseID:   "course-languages",
							CourseName: "Course Using Different Languages.",
							Role:       model.GetCourseUserRole("student"),
						},
						"course-with-lms": {
							CourseID:   "course-with-lms",
							CourseName: "Course With LMS",
							Role:       model.GetCourseUserRole("student"),
						},
						"course-without-source": {
							CourseID:   "course-without-source",
							CourseName: "Course Without Source",
							Role:       model.GetCourseUserRole("student"),
						},
						"course101": {
							CourseID:   "course101",
							CourseName: "Course 101",
							Role:       model.GetCourseUserRole("student"),
						},
						"course101-with-zero-limit": {
							CourseID:   "course101-with-zero-limit",
							CourseName: "Course 101 - With Zero Limit",
							Role:       model.GetCourseUserRole("student"),
						},
					},
				},
			},
			nil,
		},

		// Compare CourseUserInfos.
		{
			true,
			nil,
			[]*CourseUserInfo{
				{
					UserInfo: UserInfo{
						Type:  CourseUserInfoType,
						Email: "admin@test.com",
						Name:  "admin",
					},
					Role:  model.GetCourseUserRole("admin"),
					LMSID: "",
				},
				{
					UserInfo: UserInfo{
						Type:  CourseUserInfoType,
						Email: "grader@test.com",
						Name:  "grader",
					},
					Role:  model.GetCourseUserRole("grader"),
					LMSID: "",
				},
				{
					UserInfo: UserInfo{
						Type:  CourseUserInfoType,
						Email: "student@test.com",
						Name:  "student",
					},
					Role:  model.GetCourseUserRole("student"),
					LMSID: "",
				},
			},
		},
	}

	for i, testCase := range testCases {
		if !testCase.toCourseUser {
			serverUserInfos := make([]*ServerUserInfo, 0, len(serverUsers))
			for _, user := range serverUsers {
				serverUserInfos = append(serverUserInfos, MustNewServerUserInfo(user))
			}

			for _, serverUserInfo := range serverUserInfos {
				if serverUserInfo.Type != ServerUserInfoType {
					test.Errorf("Test %d: Unexpected user info type. Expected: '%s', actual: '%s'.", i, ServerUserInfoType, serverUserInfo.Type)
					continue
				}
			}

			slices.SortFunc(serverUserInfos, CompareServerUserInfoPointer)

			if !reflect.DeepEqual(testCase.serverExpected, serverUserInfos) {
				test.Errorf("Test %d: Unexpected result. Expected: '%s', actual: '%s'.",
					i, util.MustToJSONIndent(testCase.serverExpected), util.MustToJSONIndent(serverUserInfos))
				continue
			}
		} else {
			courseUsers := make([]*model.CourseUser, 0, len(serverUsers))
			for _, user := range serverUsers {
				courseUser, err := user.ToCourseUser(db.TEST_COURSE_ID, true)
				if err != nil {
					test.Errorf("Test %d: Failed to convert a server user into a course user.", i)
					continue
				}

				courseUsers = append(courseUsers, courseUser)
			}

			courseUserInfos := make([]*CourseUserInfo, 0, len(courseUsers))
			for _, courseUser := range courseUsers {
				courseUserInfos = append(courseUserInfos, NewCourseUserInfo(courseUser))
			}

			for _, courseUserInfo := range courseUserInfos {
				if courseUserInfo.Type != CourseUserInfoType {
					test.Errorf("Test %d: Unexpected user info type. Expected: '%s', actual: '%s'.", i, CourseUserInfoType, courseUserInfo.Type)
					continue
				}
			}

			slices.SortFunc(courseUserInfos, CompareCourseUserInfoPointer)

			if !reflect.DeepEqual(testCase.courseExpected, courseUserInfos) {
				test.Errorf("Test %d: Unexpected result. Expected: '%s', actual: '%s'.",
					i, util.MustToJSONIndent(testCase.courseExpected), util.MustToJSONIndent(courseUserInfos))
				continue
			}
		}
	}
}
