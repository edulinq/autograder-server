package core

import (
	"testing"

	"github.com/edulinq/autograder/internal/model"
)

func TestGetMaxServerRole(test *testing.T) {
	testCases := []struct {
		value    any
		expected model.ServerUserRole
	}{
		{struct{}{}, model.ServerRoleUnknown},
		{struct{ int }{}, model.ServerRoleUnknown},

		{struct{ MinServerRoleRoot }{}, model.ServerRoleRoot},
		{struct{ MinServerRoleOwner }{}, model.ServerRoleOwner},
		{struct{ MinServerRoleAdmin }{}, model.ServerRoleAdmin},
		{struct{ MinServerRoleCourseCreator }{}, model.ServerRoleCourseCreator},
		{struct{ MinServerRoleUser }{}, model.ServerRoleUser},

		{struct {
			MinServerRoleRoot
			MinServerRoleOwner
		}{}, model.ServerRoleRoot},
		{struct {
			MinServerRoleRoot
			MinServerRoleUser
		}{}, model.ServerRoleRoot},
		{struct {
			MinServerRoleAdmin
			MinServerRoleUser
		}{}, model.ServerRoleAdmin},

		{struct {
			A MinServerRoleAdmin
			B MinServerRoleOwner
		}{}, model.ServerRoleOwner},
		{struct {
			A MinServerRoleAdmin
			B MinServerRoleAdmin
		}{}, model.ServerRoleAdmin},
	}

	for i, testCase := range testCases {
		role, hasRole := getMaxServerRole(testCase.value)

		if testCase.expected == model.ServerRoleUnknown {
			if hasRole {
				test.Errorf("Case %d: Found a role ('%s') when none was specified.", i, role)
			}

			continue
		}

		if role != testCase.expected {
			test.Errorf("Case %d: Role mismatch. Expected: '%s', Actual: '%s'.", i, testCase.expected, role)
		}
	}
}

func TestGetMaxCourseRole(test *testing.T) {
	testCases := []struct {
		value    any
		expected model.CourseUserRole
	}{
		{struct{}{}, model.CourseRoleUnknown},
		{struct{ int }{}, model.CourseRoleUnknown},

		{struct{ MinCourseRoleOwner }{}, model.CourseRoleOwner},
		{struct{ MinCourseRoleAdmin }{}, model.CourseRoleAdmin},
		{struct{ MinCourseRoleGrader }{}, model.CourseRoleGrader},
		{struct{ MinCourseRoleStudent }{}, model.CourseRoleStudent},
		{struct{ MinCourseRoleOther }{}, model.CourseRoleOther},

		{struct {
			MinCourseRoleOwner
			MinCourseRoleOther
		}{}, model.CourseRoleOwner},
		{struct {
			MinCourseRoleAdmin
			MinCourseRoleOther
		}{}, model.CourseRoleAdmin},
		{struct {
			MinCourseRoleGrader
			MinCourseRoleOther
		}{}, model.CourseRoleGrader},
		{struct {
			MinCourseRoleStudent
			MinCourseRoleOther
		}{}, model.CourseRoleStudent},

		{struct {
			MinCourseRoleOther
			MinCourseRoleOwner
		}{}, model.CourseRoleOwner},
		{struct {
			MinCourseRoleOther
			MinCourseRoleAdmin
		}{}, model.CourseRoleAdmin},
		{struct {
			MinCourseRoleOther
			MinCourseRoleGrader
		}{}, model.CourseRoleGrader},
		{struct {
			MinCourseRoleOther
			MinCourseRoleStudent
		}{}, model.CourseRoleStudent},
	}

	for i, testCase := range testCases {
		role, hasRole := getMaxCourseRole(testCase.value)

		if testCase.expected == model.CourseRoleUnknown {
			if hasRole {
				test.Errorf("Case %d: Found a role ('%s') when none was specified.", i, role)
			}

			continue
		}

		if role != testCase.expected {
			test.Errorf("Case %d: Role mismatch. Expected: '%s', Actual: '%s'.", i, testCase.expected, role)
		}
	}
}
