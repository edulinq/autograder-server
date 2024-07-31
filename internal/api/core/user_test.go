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
		serverUserB := users[testCase.b]

		serverUserInfoA := MustNewServerUserInfo(serverUserA)
		if serverUserInfoA.GetType() != ServerUserInfoType {
			test.Errorf("Test %d: Unexpected user info type. Expected: '%s', actual: '%s'.", i, ServerUserInfoType, serverUserInfoA.GetType())
			continue
		}

		serverUserInfoB := MustNewServerUserInfo(serverUserB)
		if serverUserInfoB.GetType() != ServerUserInfoType {
			test.Errorf("Test %d: Unexpected user info type. Expected: '%s', actual: '%s'.", i, ServerUserInfoType, serverUserInfoB.GetType())
			continue
		}

		result := CompareServerUserInfoPointer(serverUserInfoA, serverUserInfoB)
		if result != testCase.expected {
			test.Errorf("Test %d: Unable to compare a serverUser with a courseUser. ServerInfoA: '%v', CourseInfoA: '%v'.", i, serverUserInfoA, serverUserInfoB)
			continue
		}

		courseUserA, err := serverUserA.ToCourseUser(db.TEST_COURSE_ID)
		if err != nil {
			test.Errorf("Test %d: Failed to convert server user A into a course user.", i)
			continue
		}

		courseUserInfoA := NewCourseUserInfo(courseUserA)
		if courseUserInfoA.GetType() != CourseUserInfoType {
			test.Errorf("Test %d: Unexpected user info type. Expected: '%s', actual: '%s'.", i, CourseUserInfoType, courseUserInfoA.GetType())
			continue
		}

		courseUserB, err := serverUserB.ToCourseUser(db.TEST_COURSE_ID)
		if err != nil {
			test.Errorf("Test %d: Failed to convert server user B into a course user.", i)
			continue
		}

		courseUserInfoB := NewCourseUserInfo(courseUserB)
		if courseUserInfoB.GetType() != CourseUserInfoType {
			test.Errorf("Test %d: Unexpected user info type. Expected: '%s', actual: '%s'.", i, CourseUserInfoType, courseUserInfoB.GetType())
			continue
		}

		result = CompareCourseUserInfoPointer(courseUserInfoA, courseUserInfoB)
		if result != testCase.expected {
			test.Errorf("Test %d: Unable to compare a serverUser with a courseUser. ServerInfoA: '%v', CourseInfoA: '%v'.", i, courseUserInfoA, courseUserInfoB)
			continue
		}
	}
}
