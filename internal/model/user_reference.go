package model

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Course user references can be represented as follows:
// - An email address.
// - A literal "*" (which includes all users in the course).
// - A course role (which will include all course users with that role).
// - Any of the above options preceded by a dash ("-") (which indicates that the user or group will NOT be included in the final results).
type CourseUserReference string

type ResolvedCourseUserReference struct {
	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of course roles to include.
	CourseUserRoles map[string]any

	// The set of course roles to exclude.
	ExcludeCourseUserRoles map[string]any
}

func (this ResolvedCourseUserReference) RefersTo(email string, role string) bool {
	_, ok := this.ExcludeEmails[email]
	if ok {
		return false
	}

	_, ok = this.ExcludeCourseUserRoles[role]
	if ok {
		return false
	}

	_, ok = this.Emails[email]
	if ok {
		return true
	}

	_, ok = this.CourseUserRoles[role]
	if ok {
		return true
	}

	return false
}

// Process a list of user inputs in the context of a course.
// See CourseUserReference for the list of acceptable inputs.
// Inputs are parsed without checking permissions.
// APIs using this function should require a sufficient role or custom permissions checking.
// Returns a reference and user errors.
// User-level errors return (partial reference, user errors).
func ResolveCourseUserReferences(rawReferences []CourseUserReference) (*ResolvedCourseUserReference, error) {
	courseUserReference := ResolvedCourseUserReference{
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		CourseUserRoles:        make(map[string]any, 0),
		ExcludeCourseUserRoles: make(map[string]any, 0),
	}

	var errs error = nil

	commonCourseRoles := GetCommonCourseUserRoleStrings()

	for _, rawReference := range rawReferences {
		reference := strings.ToLower(strings.TrimSpace(string(rawReference)))

		if reference == "" {
			continue
		}

		exclude := false
		if strings.HasPrefix(reference, "-") {
			exclude = true

			reference = strings.TrimPrefix(reference, "-")
			reference = strings.TrimSpace(reference)
		}

		if strings.Contains(reference, "@") {
			if exclude {
				courseUserReference.ExcludeEmails[reference] = nil
			} else {
				courseUserReference.Emails[reference] = nil
			}
		} else if reference == "*" {
			if exclude {
				courseUserReference.ExcludeCourseUserRoles = commonCourseRoles
			} else {
				courseUserReference.CourseUserRoles = commonCourseRoles
			}
		} else {
			// Target a specific course role.
			_, ok := commonCourseRoles[reference]
			if !ok {
				errs = errors.Join(errs, fmt.Errorf("Unknown course role '%s' in reference: '%s'.", reference, rawReference))
				continue
			}

			if exclude {
				courseUserReference.ExcludeCourseUserRoles[reference] = nil
			} else {
				courseUserReference.CourseUserRoles[reference] = nil
			}
		}
	}

	return &courseUserReference, errs
}

// Returns a sorted list of users based on the course reference.
func ResolveCourseUsers(users map[string]*CourseUser, reference *ResolvedCourseUserReference) []*CourseUser {
	if reference == nil {
		return nil
	}

	results := make([]*CourseUser, 0, len(users))

	for email, user := range users {
		if reference.RefersTo(email, user.Role.String()) {
			results = append(results, user)
		}
	}

	slices.SortFunc(results, CompareCourseUserPointer)

	return results
}

// Returns a sorted list of emails based on the course reference.
func ResolveCourseUserEmails(users map[string]*CourseUser, reference *ResolvedCourseUserReference) []string {
	if reference == nil {
		return nil
	}

	results := make([]string, 0, len(users))

	for email, user := range users {
		if reference.RefersTo(email, user.Role.String()) {
			results = append(results, email)
		}
	}

	slices.Sort(results)

	return results
}
