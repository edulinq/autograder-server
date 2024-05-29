package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Course user roles represent a user's role within a single course.
type CourseUserRole int

// RoleUnknown is the zero value and no user should have this role (it is a validation error).
// RoleOther is for miscellaneous users that should not be able to submit.
// RoleStudent is for standard users/students.
// RoleGrader is for users that need access to grades/submissions, but cannot administrate a course.
// RoleAdmin is for users that need to administrate a course.
// RoleOwner is for the top-level authorities of a course.
const (
	RoleUnknown CourseUserRole = 0
	RoleOther                  = 10
	RoleStudent                = 20
	RoleGrader                 = 30
	RoleAdmin                  = 40
	RoleOwner                  = 50
)

var courseRoleToString = map[CourseUserRole]string{
	RoleUnknown: "unknown",
	RoleOther:   "other",
	RoleStudent: "student",
	RoleGrader:  "grader",
	RoleAdmin:   "admin",
	RoleOwner:   "owner",
}

var stringToCourseUserRole = map[string]CourseUserRole{
	"unknown": RoleUnknown,
	"other":   RoleOther,
	"student": RoleStudent,
	"grader":  RoleGrader,
	"admin":   RoleAdmin,
	"owner":   RoleOwner,
}

func GetCourseUserRole(text string) CourseUserRole {
	return stringToCourseUserRole[text]
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
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(courseRoleToString[this])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (this *CourseUserRole) UnmarshalJSON(data []byte) error {
	var temp string

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	temp = strings.ToLower(temp)

	var ok bool
	*this, ok = stringToCourseUserRole[temp]
	if !ok {
		*this = RoleUnknown
		return fmt.Errorf("RoleUnknown CourseUserRole value: '%s'.", temp)
	}

	return nil
}
