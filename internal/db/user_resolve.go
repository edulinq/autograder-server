package db

import (
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// Maps string representations of roles and * (all roles) to the emails for users with those roles.
// The function takes a course and a list of strings, containing emails, roles, and * as input and returns a sorted slice of lowercase emails without duplicates.
func ResolveCourseUsers(course *model.Course, emails []string) ([]string, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	emailSet := map[string]any{}
	roleSet := map[string]any{}

	// Iterate over all strings, checking for emails, roles, and * (which denotes all users).
	for _, email := range emails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}

		if strings.Contains(email, "@") {
			emailSet[email] = nil
		} else {
			if email == "*" {
				allRoles := model.GetAllRoleStrings()
				for role := range allRoles {
					roleSet[role] = nil
				}
			} else {
				if model.GetRole(email) == model.RoleUnknown {
					log.Warn("Invalid role, cannot resolve users.", course, log.NewAttr("role", email))
					continue
				}

				roleSet[email] = nil
			}
		}
	}

	if len(roleSet) > 0 {
		users, err := GetCourseUsers(course)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			// Add a user if their role is set.
			_, ok := roleSet[model.GetRoleString(user.Role)]
			if ok {
				emailSet[strings.ToLower(user.Email)] = nil
			}
		}
	}

	emailSlice := make([]string, 0, len(emailSet))
	for email := range emailSet {
		emailSlice = append(emailSlice, email)
	}

	slices.Sort(emailSlice)

	return emailSlice, nil
}
