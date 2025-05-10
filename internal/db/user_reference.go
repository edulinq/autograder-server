package db

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

const USER_REFERENCE_DELIM = "::"

const (
	EmailReference      UserReferenceType = 0
	CourseRoleReference                   = 10
	CourseReference                       = 20
	ServerRoleReference                   = 30
	AllUserReference                      = 40
)

type UserReferenceType int

// A flexible way to reference server users.
// Server user references can be represented as follows (in the order the are evaluated):
//
// - An email address
// - An email address preceded by a dash ("-") (which indicates that this email address should NOT be included in the final results).
// - A server role (which will include all server users with that role)
// - A literal "*" (which includes all users on the server)
// TODO: Update final part of this comment (show examples of various forms)
// - A course user reference
type ServerUserReference string

// Course user references can be represented as follows (in the order they are evaluated):
//
// - An email address
// - An email address preceded by a dash ("-") (which indicates that this email address should NOT be included in the final results).
// - A course role (which will include all course users with that role)
// - A literal "*" (which includes all users in the course)
// - A course user reference
type CourseUserReference string

type UserReference struct {
	// The type of the user reference.
	Type UserReferenceType

	// A signal to exclude the users captured in the reference.
	Exclude bool

	// The email address of the user.
	Email string

	// Refers to all users with this server role.
	ServerUserRole model.ServerUserRole

	// The course that orients the user reference.
	// If there is a course but not a course role,
	// refers to all users in the course.
	// If the course is nil but there is a course role,
	// refers to all users with the target course role in ANY course.
	Course *model.Course

	CourseUserRole model.CourseUserRole
}

func ParseUserReference(rawReference string) (*UserReference, error) {
	reference := strings.ToLower(strings.TrimSpace(rawReference))

	exclude := false
	if strings.HasPrefix(reference, "-") {
		exclude = true

		reference = strings.TrimPrefix(reference, "-")
	}

	if strings.Contains(reference, "@") {
		return &UserReference{
			Type:    EmailReference,
			Exclude: exclude,
			Email:   reference,
		}, nil
	}

	if reference == "root" {
		return nil, fmt.Errorf("User reference cannot target the root user: '%s'.", rawReference)
	}

	if reference == "*" {
		return &UserReference{
			Type:    AllUserReference,
			Exclude: exclude,
		}, nil
	}

	referenceParts := strings.Split(reference, USER_REFERENCE_DELIM)
	if len(referenceParts) == 1 {
		serverRole := model.GetServerUserRole(reference)
		if serverRole != model.ServerRoleUnknown {
			return &UserReference{
				Type:           ServerRoleReference,
				Exclude:        exclude,
				ServerUserRole: serverRole,
			}, nil
		}

		course, err := GetCourse(reference)
		if err != nil {
			return nil, fmt.Errorf("Unable to get course while parsing user reference: '%w'.", err)
		}

		if course != nil {
			return &UserReference{
				Type:    CourseReference,
				Exclude: exclude,
				Course:  course,
			}, nil
		}

		return nil, fmt.Errorf("Unknown user reference: '%s'.", rawReference)
	} else if len(referenceParts) == 2 {
		if referenceParts[0] == "*" && referenceParts[1] == "*" {
			return &UserReference{
				Type:    AllUserReference,
				Exclude: exclude,
			}, nil
		}

		// Reference to all users with a certain course role.
		if referenceParts[0] == "*" {
			courseRole := model.GetCourseUserRole(referenceParts[1])
			if courseRole != model.CourseRoleUnknown {
				return &UserReference{
					Type:           CourseRoleReference,
					Course:         nil,
					CourseUserRole: courseRole,
					Exclude:        exclude,
				}, nil
			}

			return nil, fmt.Errorf("Unknown user reference: '%s'.", rawReference)
		}

		// First part must be a course.
		course, err := GetCourse(reference)
		if err != nil {
			return nil, fmt.Errorf("Unable to get course while parsing user reference: '%w'.", err)
		}

		if course == nil {
			return nil, fmt.Errorf("Unknown user reference: '%s'.", rawReference)
		}

		if referenceParts[1] == "*" {
			return &UserReference{
				Type:    CourseReference,
				Course:  course,
				Exclude: exclude,
			}, nil
		}

		courseRole := model.GetCourseUserRole(referenceParts[1])
		if courseRole == model.CourseRoleUnknown {
			return nil, fmt.Errorf("Unknown user reference: '%s'.", rawReference)
		}

		return &UserReference{
			Type:           CourseRoleReference,
			Course:         course,
			CourseUserRole: courseRole,
			Exclude:        exclude,
		}, nil
	} else {
		return nil, fmt.Errorf("Invalid user reference format: '%s'.", rawReference)
	}
}

func ParseCourseUserReference(course *model.Course, rawReference string) (*UserReference, error) {
	reference := strings.ToLower(strings.TrimSpace(rawReference))

	exclude := false
	if strings.HasPrefix(reference, "-") {
		exclude = true

		reference = strings.TrimPrefix(reference, "-")
	}

	userReference := &UserReference{
		Course:  course,
		Exclude: exclude,
	}

	if strings.Contains(reference, "@") {
		userReference.Type = EmailReference
		userReference.Email = reference

		return userReference, nil
	}

	if reference == "root" {
		return nil, fmt.Errorf("User reference cannot target the root user: '%s'.", rawReference)
	}

	if reference == "*" {
		userReference.Type = CourseReference

		return userReference, nil
	}

	courseRole := model.GetCourseUserRole(reference)
	if courseRole != model.CourseRoleUnknown {
		userReference.Type = CourseRoleReference
		userReference.CourseUserRole = courseRole

		return userReference, nil
	}

	return nil, fmt.Errorf("Invalid course user reference: '%s'.", rawReference)
}
