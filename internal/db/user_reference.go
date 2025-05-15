package db

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

// A flexible way to reference server users.
// Server user references can be represented as follows (in the order the are evaluated):
//
// - An email address
// - An email address preceded by a dash ("-") (which indicates that this email address should NOT be included in the final results).
// - A server role (which will include all server users with that role)
// - A literal "*" (which includes all users on the server)
// TODO: Update final part of this comment (show examples of various forms)
// - A course user reference
type ServerUserReferenceInput string

// Course user references can be represented as follows (in the order they are evaluated):
//
// - An email address
// - An email address preceded by a dash ("-") (which indicates that this email address should NOT be included in the final results).
// - A course role (which will include all course users with that role)
// - A literal "*" (which includes all users in the course)
// - A course user reference
type CourseUserReferenceInput string

func ParseUserReference(rawReferences []ServerUserReferenceInput) (*model.ServerUserReference, error) {
	serverUserReference := &model.ServerUserReference{
		Emails:                  make(map[string]any, 0),
		ExcludeEmails:           make(map[string]any, 0),
		ServerUserRoles:         make(map[model.ServerUserRole]any, 0),
		ExcludeServerUserRoles:  make(map[model.ServerUserRole]any, 0),
		CourseReferences:        make(map[string]CourseUserReference, 0),
		ExcludeCourseReferences: make(map[string]any, 0),
	}

	var errs error = nil

	for i, rawReference := range rawReferences {
		reference := strings.ToLower(strings.TrimSpace(string(rawReference)))

		exclude := false
		if strings.HasPrefix(reference, "-") {
			exclude = true

			reference = strings.TrimPrefix(reference, "-")
		}

		if reference == "root" {
			errs = errors.Join(errs, fmt.Errorf("User reference %d cannot target the root user: '%s'.", i, rawReference))
			continue
		}

		if strings.Contains(reference, "@") {
			if exclude {
				serverUserReference.ExcludeEmails[reference] = nil
			} else {
				serverUserReference.Emails[reference] = nil
			}

			continue
		}

		if reference == "*" {
			serverUserReference.AllUsers = true
			continue
		}

		referenceParts := strings.Split(reference, USER_REFERENCE_DELIM)
		if len(referenceParts) == 1 {
			serverRole := model.GetServerUserRole(reference)
			if serverRole != model.ServerRoleUnknown {
				if exclude {
					serverUserReference.ExclueServerUserRoles[serverRole] = nil
				} else {
					serverUserReference.ServerUserRoles[serverRole] = nil
				}

				continue
			}

			errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'.", i, rawReference))
		} else if len(referenceParts) == 2 {
			if referenceParts[0] == "*" && referenceParts[1] == "*" {
				userReference.AllUsers = true
				continue
			}

			// TODO: Continue from here!
			// Reference to all users with a certain course role.
			if referenceParts[0] == "*" {
				courseRole := model.GetCourseUserRole(referenceParts[1])
				if courseRole != model.CourseRoleUnknown {
					return &model.ServerUserReference{
						Course:         nil,
						CourseUserRole: courseRole,
						Exclude:        exclude,
					}, nil
				}

				return nil, fmt.Errorf("Unknown user reference: '%s'.", rawReference)
			}

			// First part must be a course.
			course, err := GetCourse(referenceParts[0])
			if err != nil {
				return nil, fmt.Errorf("Unable to get course while parsing user reference: '%w'.", err)
			}

			if course == nil {
				return nil, fmt.Errorf("Unknown user reference: '%s'.", rawReference)
			}

			if referenceParts[1] == "*" {
				return &model.ServerUserReference{
					Course:  course,
					Exclude: exclude,
				}, nil
			}

			courseRole := model.GetCourseUserRole(referenceParts[1])
			if courseRole == model.CourseRoleUnknown {
				return nil, fmt.Errorf("Unknown user reference: '%s'.", rawReference)
			}

			return &model.ServerUserReference{
				Course:         course,
				CourseUserRole: courseRole,
				Exclude:        exclude,
			}, nil
		} else {
			return nil, fmt.Errorf("Invalid user reference format: '%s'.", rawReference)
		}

	}
}

func ParseCourseUserReference(course *model.Course, rawReference CourseUserReferenceInput) (*CourseUserReference, error) {
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
		userReference.Email = reference

		return userReference, nil
	}

	if reference == "root" {
		return nil, fmt.Errorf("User reference cannot target the root user: '%s'.", rawReference)
	}

	if reference == "*" {

		return userReference, nil
	}

	courseRole := model.GetCourseUserRole(reference)
	if courseRole != model.CourseRoleUnknown {
		userReference.CourseUserRole = courseRole

		return userReference, nil
	}

	return nil, fmt.Errorf("Invalid course user reference: '%s'.", rawReference)
}
