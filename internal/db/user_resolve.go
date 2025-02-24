package db

import (
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// Resolve course email addresses.
// Take a course and a list of strings (containing emails specs) and returns a sorted slice of lowercase emails without duplicates.
// An email spec can be:
// an email address,
// a course role (which will include all course users with that role),
// a literal "*" (which includes all users enrolled in the course),
// or an email address preceded by a dash ("-") (which indicates that this email address should NOT be included in the final results).
func ResolveCourseUsers(course *model.Course, emails []string) ([]string, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	emailSet := map[string]any{}
	removeEmailSet := map[string]any{}
	roleSet := map[string]any{}

	// Iterate over all strings, checking for emails, roles, and * (which denotes all users).
	for _, email := range emails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}

		if strings.HasPrefix(email, "-") {
			email = strings.TrimSpace(strings.TrimPrefix(email, "-"))
			removeEmailSet[email] = nil
		} else if strings.Contains(email, "@") {
			emailSet[email] = nil
		} else {
			if email == "*" {
				allRoles := model.GetAllCourseUserRolesStrings()
				for role := range allRoles {
					roleSet[role] = nil
				}
			} else {
				if model.GetCourseUserRole(email) == model.CourseRoleUnknown {
					log.Warn("Invalid role, cannot resolve users.", course, log.NewAttr("role", email))
					continue
				}

				roleSet[email] = nil
			}
		}
	}

	// Add users from roles.
	if len(roleSet) > 0 {
		users, err := GetCourseUsers(course)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			// Add a user if their role is set.
			_, ok := roleSet[user.Role.String()]
			if ok {
				emailSet[strings.ToLower(user.Email)] = nil
			}
		}
	}

	// Remove negative users.
	for email, _ := range removeEmailSet {
		delete(emailSet, email)
	}

	emailSlice := make([]string, 0, len(emailSet))
	for email := range emailSet {
		emailSlice = append(emailSlice, email)
	}

	slices.Sort(emailSlice)

	return emailSlice, nil
}
