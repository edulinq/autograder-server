package model

import (
	"fmt"
	"slices"
	"strings"
)

// Course user references can represent the following:
// - An email address.
// - A literal "*" (which includes all users in the course).
// - A course role (which will include all course users with that role).
// - Any of the above options preceded by a dash ("-") (which indicates that the user or group will NOT be included in the final results).
type CourseUserReference string

type ParsedCourseUserReference struct {
	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of course roles to include.
	CourseUserRoles map[CourseUserRole]any

	// The set of course roles to exclude.
	ExcludeCourseUserRoles map[CourseUserRole]any
}

func (this ParsedCourseUserReference) RefersTo(email string, role CourseUserRole) bool {
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

// Parse a list of user inputs into a structured reference.
// See CourseUserReference for the list of acceptable inputs.
// Inputs are parsed without checking permissions.
// APIs using this function should require a sufficient role or custom permissions checking.
// Returns a reference and user errors.
// User-level errors return (partial reference, user errors).
func ParseCourseUserReferences(rawReferences []CourseUserReference) (*ParsedCourseUserReference, map[string]error) {
	courseUserReference := ParsedCourseUserReference{
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		CourseUserRoles:        make(map[CourseUserRole]any, 0),
		ExcludeCourseUserRoles: make(map[CourseUserRole]any, 0),
	}

	userErrors := make(map[string]error, 0)

	commonCourseRoles := make(map[string]CourseUserRole, len(CommonCourseUserRole))
	for roleString, role := range CommonCourseUserRole {
		commonCourseRoles[roleString] = role
	}

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
			allCourseRoles := make(map[CourseUserRole]any, len(commonCourseRoles))
			for _, role := range commonCourseRoles {
				allCourseRoles[role] = nil
			}

			if exclude {
				courseUserReference.ExcludeCourseUserRoles = allCourseRoles
			} else {
				courseUserReference.CourseUserRoles = allCourseRoles
			}
		} else {
			// Target a specific course role.
			role, ok := commonCourseRoles[reference]
			if !ok {
				userErrors[string(rawReference)] = fmt.Errorf("Unknown course role '%s'.", reference)
				continue
			}

			if exclude {
				courseUserReference.ExcludeCourseUserRoles[role] = nil
			} else {
				courseUserReference.CourseUserRoles[role] = nil
			}
		}
	}

	if len(userErrors) == 0 {
		userErrors = nil
	}

	return &courseUserReference, userErrors
}

// Returns a sorted list of users based on the course reference.
func ResolveCourseUsers(users map[string]*CourseUser, reference *ParsedCourseUserReference) []*CourseUser {
	if reference == nil {
		return nil
	}

	results := make([]*CourseUser, 0, len(users))

	for email, user := range users {
		if reference.RefersTo(email, user.Role) {
			results = append(results, user)
		}
	}

	slices.SortFunc(results, CompareCourseUserPointer)

	return results
}

// Returns a sorted list of emails based on the course reference.
// Emails can target users outside of the course.
func ResolveCourseUserEmails(users map[string]*CourseUser, reference *ParsedCourseUserReference) []string {
	if reference == nil {
		return nil
	}

	emailSet := make(map[string]any, 0)
	excludeSet := make(map[string]any, 0)

	// Add all emails from the course users.
	for email, user := range users {
		if reference.RefersTo(email, user.Role) {
			emailSet[email] = nil
		} else {
			excludeSet[email] = nil
		}
	}

	// Add all emails based on email alone.
	for email, _ := range reference.Emails {
		_, ok := reference.ExcludeEmails[email]
		if ok {
			continue
		}

		_, ok = excludeSet[email]
		if ok {
			continue
		}

		emailSet[email] = nil
	}

	results := make([]string, 0, len(emailSet))

	for email, _ := range emailSet {
		results = append(results, email)
	}

	slices.Sort(results)

	return results
}
