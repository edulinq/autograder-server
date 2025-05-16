package db

import (
	"errors"
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
		CourseReferences:        make(map[string]model.CourseUserReference, 0),
		ExcludeCourseReferences: make(map[string]any, 0),
	}

	var errs error = nil
	var err error = nil

	for i, rawReference := range rawReferences {
		reference := strings.ToLower(strings.TrimSpace(string(rawReference)))

		exclude := false
		if strings.HasPrefix(reference, "-") {
			exclude = true

			reference = strings.TrimPrefix(reference, "-")
		}

		// Only process exclusions after targeting all users.
		if serverUserReference.AllUsers && !exclude {
			continue
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

		parts := strings.Split(reference, model.USER_REFERENCE_DELIM)
		if len(parts) == 1 {
			serverRole := model.GetServerUserRole(reference)
			if serverRole == model.ServerRoleUnknown {
				errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'. Unknown server user role '%s'.", i, rawReference, reference))

				continue
			}

			if exclude {
				serverUserReference.ExcludeServerUserRoles[serverRole] = nil
			} else {
				serverUserReference.ServerUserRoles[serverRole] = nil
			}
		} else if len(parts) == 2 {
			if parts[0] == "*" && parts[1] == "*" {
				serverUserReference.AllUsers = true
				continue
			}

			// TODO: Can we leverage the parsing function of the course user reference?
			// We could get back a CourseUserReference and then merge it in?
			courses := make(map[string]*model.Course, 0)

			// Target all courses.
			if parts[0] == "*" {
				// TODO: Think about caching all courses. (or implementing db caching).
				courses, err = GetCourses()
				if err != nil {
					return nil, fmt.Errorf("Failed to get courses: '%w'.", err)
				}
			} else {
				// Target a specific course.
				course, err := GetCourse(parts[0])
				if err != nil {
					errs = errors.Join(errs, fmt.Errorf("Unable to get course while parsing user reference: '%w'.", err))

					continue
				}

				if course == nil {
					errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'. Unknown course '%s'.", i, rawReference, parts[0]))

					continue
				}

				courses[course.GetID()] = course
			}

			courseRoles := make([]model.CourseUserRole, 0)

			// Target all course roles.
			if parts[1] == "*" {
				courseRoles := model.GetCommonCourseRoleStrings()
			} else {
				// Target a specific course role.
				courseRole := model.GetCourseUserRole(parts[1])
				if courseRole == model.CourseRoleUnknown {
					errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'. Unknown course user role '%s'.", i, rawReference, parts[1]))

					continue
				}

				courseRoles = append(courseRoles, courseRole)
			}

			serverUserReference.AddCourseUserReference(courses, courseRoles, exclude)
		} else {
			return nil, fmt.Errorf("Invalid user reference format: '%s'.", rawReference)
		}
	}

	return serverUserReference, errs
}

// TODO: Convert course function.
func ParseCourseUserReference(course *model.Course, rawReference CourseUserReferenceInput) (*model.CourseUserReference, error) {
	reference := strings.ToLower(strings.TrimSpace(rawReference))

	exclude := false
	if strings.HasPrefix(reference, "-") {
		exclude = true

		reference = strings.TrimPrefix(reference, "-")
	}

	userReference := &model.CourseUserReference{
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
