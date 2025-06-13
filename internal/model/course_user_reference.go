package model

import (
	"errors"
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

func (this *ParsedCourseUserReference) ToParsedServerUserReference(courseID string) *ParsedServerUserReference {
	if this == nil {
		return nil
	}

	return &ParsedServerUserReference{
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		ServerUserRoles:        make(map[ServerUserRole]any, 0),
		ExcludeServerUserRoles: make(map[ServerUserRole]any, 0),
		CourseUserReferences: map[string]*ParsedCourseUserReference{
			courseID: this,
		},
	}
}

func (this *ParsedCourseUserReference) Merge(other *ParsedCourseUserReference) *ParsedCourseUserReference {
	if this == other {
		return this
	}

	if this == nil {
		return other
	}

	if other == nil {
		return this
	}

	for email, _ := range other.Emails {
		this.Emails[email] = nil
	}

	for email, _ := range other.ExcludeEmails {
		this.ExcludeEmails[email] = nil
	}

	for role, _ := range other.CourseUserRoles {
		this.CourseUserRoles[role] = nil
	}

	for role, _ := range other.ExcludeCourseUserRoles {
		this.ExcludeCourseUserRoles[role] = nil
	}

	return this
}

func (this ParsedCourseUserReference) Excludes(email string, role CourseUserRole) bool {
	_, ok := this.ExcludeEmails[email]
	if ok {
		return true
	}

	_, ok = this.ExcludeCourseUserRoles[role]
	if ok {
		return true
	}

	return false
}

func (this ParsedCourseUserReference) RefersTo(email string, role CourseUserRole) bool {
	if this.Excludes(email, role) {
		return false
	}

	_, ok := this.Emails[email]
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
func ParseCourseUserReferences(rawReferences []CourseUserReference) (*ParsedCourseUserReference, error) {
	courseUserReference := ParsedCourseUserReference{
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		CourseUserRoles:        make(map[CourseUserRole]any, 0),
		ExcludeCourseUserRoles: make(map[CourseUserRole]any, 0),
	}

	var errs error = nil

	commonCourseRoles := GetCommonCourseUserRolesCopy()

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
				errs = errors.Join(errs, fmt.Errorf("Course user reference '%s' contains an unknown course role: '%s'.", rawReference, reference))
				continue
			}

			if exclude {
				courseUserReference.ExcludeCourseUserRoles[role] = nil
			} else {
				courseUserReference.CourseUserRoles[role] = nil
			}
		}
	}

	return &courseUserReference, errs
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
	// Exclusion always takes priority over inclusion.
	// The final list of emails will be the users in emailSet that are not in excludeSet.
	// (e.g., a user excluded based on role but included by explicit email will NOT be included in the results).
	excludeSet := make(map[string]any, 0)

	// Add all emails from the course users.
	for email, user := range users {
		if reference.Excludes(email, user.Role) {
			excludeSet[email] = nil
		} else if reference.RefersTo(email, user.Role) {
			emailSet[email] = nil
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

func createCourseUserReference(courseRoles map[CourseUserRole]any, exclude bool) *ParsedCourseUserReference {
	courseUserRoles := make(map[CourseUserRole]any, 0)
	excludeCourseUserRoles := make(map[CourseUserRole]any, 0)

	if exclude {
		excludeCourseUserRoles = courseRoles
	} else {
		courseUserRoles = courseRoles
	}

	return &ParsedCourseUserReference{
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		CourseUserRoles:        courseUserRoles,
		ExcludeCourseUserRoles: excludeCourseUserRoles,
	}
}
