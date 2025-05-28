package model

import (
	"github.com/edulinq/autograder/internal/util"
)

// Course user roles represent a user's role within a single course.
type CourseUserRole int

// CourseRoleUnknown is the zero value and no user should have this role (it is a validation error).
// CourseRoleOther is for miscellaneous users that should not be able to submit.
// CourseRoleStudent is for standard users/students.
// CourseRoleGrader is for users that need access to grades/submissions, but cannot administrate a course.
// CourseRoleAdmin is for users that need to administrate a course.
// CourseRoleOwner is for the top-level authorities of a course.
const (
	CourseRoleUnknown CourseUserRole = 0
	CourseRoleOther                  = 10
	CourseRoleStudent                = 20
	CourseRoleGrader                 = 30
	CourseRoleAdmin                  = 40
	CourseRoleOwner                  = 50
)

var courseRoleToString = map[CourseUserRole]string{
	CourseRoleUnknown: "unknown",
	CourseRoleOther:   "other",
	CourseRoleStudent: "student",
	CourseRoleGrader:  "grader",
	CourseRoleAdmin:   "admin",
	CourseRoleOwner:   "owner",
}

var stringToCourseUserRole = map[string]CourseUserRole{
	"unknown": CourseRoleUnknown,
	"other":   CourseRoleOther,
	"student": CourseRoleStudent,
	"grader":  CourseRoleGrader,
	"admin":   CourseRoleAdmin,
	"owner":   CourseRoleOwner,
}

// Common course user roles are the valid roles a user could have in a course.
var CommonCourseUserRole = map[string]CourseUserRole{
	"other":   CourseRoleOther,
	"student": CourseRoleStudent,
	"grader":  CourseRoleGrader,
	"admin":   CourseRoleAdmin,
	"owner":   CourseRoleOwner,
}

func GetCourseUserRole(text string) CourseUserRole {
	return stringToCourseUserRole[text]
}

func GetCourseUserRoleString(role CourseUserRole) string {
	return courseRoleToString[role]
}

func GetAllCourseUserRoles() map[CourseUserRole]string {
	return courseRoleToString
}

func GetAllCourseUserRolesStrings() map[string]CourseUserRole {
	return stringToCourseUserRole
}

func (this CourseUserRole) String() string {
	return courseRoleToString[this]
}

func (this CourseUserRole) MarshalJSON() ([]byte, error) {
	return util.MarshalEnum(this, courseRoleToString)
}

func (this *CourseUserRole) UnmarshalJSON(data []byte) error {
	value, err := util.UnmarshalEnum(data, stringToCourseUserRole, true)
	if err == nil {
		*this = *value
	}

	return err
}
