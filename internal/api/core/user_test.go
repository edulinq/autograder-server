package core

import (
	"testing"

	"github.com/edulinq/autograder/internal/db"
)

func TestNewUserInfos(test *testing.T) {
	users := db.MustGetServerUsers()

	testCases := []struct {
		a        string
		b        string
		expected int
	}{
		// Equivalnce checks.
		{"admin@test.com", "admin@test.com", 0},
		{"student@test.com", "student@test.com", 0},

		// A > B.
		{"student@test.com", "admin@test.com", 1},
		{"grader@test.com", "admin@test.com", 1},

		// A < B.
		{"admin@test.com", "grader@test.com", -1},
		{"admin@test.com", "student@test.com", -1},
	}

	for i, testCase := range testCases {
		serverUserA := users[testCase.a]
		courseUserA, err := serverUserA.ToCourseUser(db.TEST_COURSE_ID)
		if err != nil {
			test.Errorf("Test %d: Failed to convert server user A into a course user.", i)
			continue
		}

		serverUserInfoA := MustNewServerUserInfo(serverUserA)
		if serverUserInfoA.GetType() != ServerUserInfoType {
			test.Errorf("Test %d: Unexpected user info type. Expected: '%d', actual: '%d'.", i, ServerUserInfoType, serverUserInfoA.GetType())
			continue
		}

		courseUserInfoA := NewCourseUserInfo(courseUserA)
		if courseUserInfoA.GetType() != CourseUserInfoType {
			test.Errorf("Test %d: Unexpected user info type. Expected: '%d', actual: '%d'.", i, CourseUserInfoType, courseUserInfoA.GetType())
			continue
		}

		if CompareUserInfo(serverUserInfoA.UserInfo, courseUserInfoA.UserInfo) != 0 {
			test.Errorf("Test %d: The embedded user info does not match. ServerInfoA: '%v', CourseInfoA: '%v'.", i, serverUserInfoA.UserInfo, courseUserInfoA.UserInfo)
			continue
		}

		serverUserB := users[testCase.b]
		courseUserB, err := serverUserB.ToCourseUser(db.TEST_COURSE_ID)
		if err != nil {
			test.Errorf("Test %d: Failed to convert server user B into a course user.", i)
			continue
		}

		serverUserInfoB := MustNewServerUserInfo(serverUserB)
		if serverUserInfoB.GetType() != ServerUserInfoType {
			test.Errorf("Test %d: Unexpected user info type. Expected: '%d', actual: '%d'.", i, ServerUserInfoType, serverUserInfoB.GetType())
			continue
		}

		courseUserInfoB := NewCourseUserInfo(courseUserB)
		if courseUserInfoB.GetType() != CourseUserInfoType {
			test.Errorf("Test %d: Unexpected user info type. Expected: '%d', actual: '%d'.", i, CourseUserInfoType, courseUserInfoB.GetType())
			continue
		}

		if CompareUserInfo(serverUserInfoB.UserInfo, courseUserInfoB.UserInfo) != 0 {
			test.Errorf("Test %d: The embedded user info does not match. ServerInfoB: '%v', CourseInfoB: '%v'.", i, serverUserInfoB.UserInfo, courseUserInfoB.UserInfo)
			continue
		}

		result := CompareUserInfoPointer(serverUserInfoA, serverUserInfoB)
		if result != testCase.expected {
			test.Errorf("Test %d: Unexpected server comparison result. Expected: '%d', actual: '%d'.", i, testCase.expected, result)
			continue
		}

		result = CompareCourseUserInfoPointer(courseUserInfoA, courseUserInfoB)
		if result != testCase.expected {
			test.Errorf("Test %d: Unexpected course comparison result. Expected: '%d', actual: '%d'.", i, testCase.expected, result)
			continue
		}
	}
}
