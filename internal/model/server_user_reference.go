package model

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const USER_REFERENCE_DELIM = "::"

// Server user references can represent the following:
// - An email address.
// - A literal "*" (which includes all users on the server).
// - A server role (which will include all server users with that role).
// - <course-id>::<course-role> (which will include all users in the target course with that role).
// - *::<course-role> (which will include all users with the course role in any course).
// - <course-id>::* (which will include all users in the target course).
// - *::* (which includes all users enrolled in at least one course).
// - Any of the above options preceded by a dash ("-") (which indicates that the user or group will NOT be included in the final results).
type ServerUserReference string

type ParsedServerUserReference struct {
	// The set of emails to include.
	Emails map[string]any

	// The set of emails to exclude.
	ExcludeEmails map[string]any

	// The set of server roles to include.
	ServerUserRoles map[ServerUserRole]any

	// The set of server roles to exclude.
	ExcludeServerUserRoles map[ServerUserRole]any

	// Information to include or exclude users based on course information.
	// Keyed on the course ID.
	CourseUserReferences map[string]*ParsedCourseUserReference
}

func (this *ParsedServerUserReference) AddParsedCourseUserReference(courseID string, courseUserReference *ParsedCourseUserReference) {
	if this == nil {
		this = courseUserReference.ToParsedServerUserReference(courseID)
		return
	}

	this.CourseUserReferences[courseID] = this.CourseUserReferences[courseID].Merge(courseUserReference)
}

func (this ParsedServerUserReference) Excludes(user *ServerUser) bool {
	_, ok := this.ExcludeEmails[user.Email]
	if ok {
		return true
	}

	_, ok = this.ExcludeServerUserRoles[user.Role]
	if ok {
		return true
	}

	for courseID, courseReference := range this.CourseUserReferences {
		courseInfo, ok := user.CourseInfo[courseID]
		if !ok {
			continue
		}

		if courseReference.Excludes(user.Email, courseInfo.Role) {
			return true
		}
	}

	return false
}

func (this ParsedServerUserReference) RefersTo(user *ServerUser) bool {
	if this.Excludes(user) {
		return false
	}

	_, ok := this.Emails[user.Email]
	if ok {
		return true
	}

	_, ok = this.ServerUserRoles[user.Role]
	if ok {
		return true
	}

	for courseID, courseReference := range this.CourseUserReferences {
		courseInfo, ok := user.CourseInfo[courseID]
		if !ok {
			continue
		}

		if courseReference.RefersTo(user.Email, courseInfo.Role) {
			return true
		}
	}

	return false
}

func ParseServerUserReferences(rawReferences []ServerUserReference, courses map[string]*Course) (*ParsedServerUserReference, error) {
	serverUserReference := ParsedServerUserReference{
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		ServerUserRoles:        make(map[ServerUserRole]any, 0),
		ExcludeServerUserRoles: make(map[ServerUserRole]any, 0),
		CourseUserReferences:   make(map[string]*ParsedCourseUserReference, 0),
	}

	var errs error = nil

	commonServerRoles := GetCommonServerUserRolesCopy()

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
				serverUserReference.ExcludeEmails[reference] = nil
			} else {
				serverUserReference.Emails[reference] = nil
			}
		} else if reference == "*" {
			allServerRoles := make(map[ServerUserRole]any, len(commonServerRoles))
			for _, role := range commonServerRoles {
				allServerRoles[role] = nil
			}

			if exclude {
				serverUserReference.ExcludeServerUserRoles = allServerRoles
			} else {
				serverUserReference.ServerUserRoles = allServerRoles
			}
		} else {
			parts := strings.Split(reference, USER_REFERENCE_DELIM)
			if len(parts) == 1 {
				// User reference must be a server role.
				role, ok := commonServerRoles[reference]
				if !ok {
					errs = errors.Join(errs, fmt.Errorf("Server user reference '%s' contains an unknown server role: '%s'.", rawReference, reference))
					continue
				}

				if exclude {
					serverUserReference.ExcludeServerUserRoles[role] = nil
				} else {
					serverUserReference.ServerUserRoles[role] = nil
				}
			} else if len(parts) == 2 {
				// User reference must be <course-id>::<course-role>.
				// If a '*' is present, target all courses or course roles respectively.
				courseID := strings.TrimSpace(parts[0])
				courseRoleString := strings.TrimSpace(parts[1])

				err := serverUserReference.parseCourseInformation(rawReference, exclude, courseID, courseRoleString, courses)
				if err != nil {
					errs = errors.Join(errs, err)
					continue
				}
			} else {
				errs = errors.Join(errs, fmt.Errorf("Invalid format in server user reference: '%s'.", rawReference))
				continue
			}
		}
	}

	return &serverUserReference, errs
}

func (this ParsedServerUserReference) parseCourseInformation(rawReference ServerUserReference, exclude bool, courseID string, courseRoleString string, courses map[string]*Course) error {
	targetCourses := make(map[string]*Course, 0)
	if courseID == "*" {
		// Target all courses.
		targetCourses = courses
	} else {
		// Target a specific course.
		course, ok := courses[courseID]
		if !ok {
			return fmt.Errorf("Server user reference '%s' contains an unknown course: '%s'.", rawReference, courseID)
		}

		targetCourses[course.GetID()] = course
	}

	commonCourseRoles := GetCommonCourseUserRolesCopy()

	courseRoles := make(map[CourseUserRole]any, 0)
	if courseRoleString == "*" {
		// Target all course roles.
		allCourseRoles := make(map[CourseUserRole]any, len(commonCourseRoles))
		for _, role := range commonCourseRoles {
			allCourseRoles[role] = nil
		}

		courseRoles = allCourseRoles
	} else {
		// Target a specific course role.
		role, ok := commonCourseRoles[courseRoleString]
		if !ok {
			return fmt.Errorf("Server user reference '%s' contains an unknown course role: '%s'.", rawReference, courseRoleString)
		}

		courseRoles[role] = nil
	}

	for courseID, _ := range targetCourses {
		courseUserReference := courseUserRolesToParsedCourseUserReference(courseRoles, exclude)
		this.AddParsedCourseUserReference(courseID, courseUserReference)
	}

	return nil
}

// Returns a sorted list of users based on the server reference.
func ResolveServerUsers(users map[string]*ServerUser, reference *ParsedServerUserReference) []*ServerUser {
	if reference == nil {
		return nil
	}

	results := make([]*ServerUser, 0, len(users))

	for _, user := range users {
		if reference.RefersTo(user) {
			results = append(results, user)
		}
	}

	slices.SortFunc(results, CompareServerUserPointer)

	return results
}

// Returns a sorted list of emails based on the server reference.
// Emails can target users outside of the server.
func ResolveServerUserEmails(users map[string]*ServerUser, reference *ParsedServerUserReference) []string {
	if reference == nil {
		return nil
	}

	emailSet := make(map[string]any, 0)
	// Exclusion always takes priority over inclusion.
	// The final list of emails will be the users in emailSet that are not in excludeSet.
	// (e.g., a user excluded based on role but included by explicit email will NOT be included in the results).
	excludeSet := make(map[string]any, 0)

	// Add all emails from the server users.
	for email, user := range users {
		if reference.Excludes(user) {
			excludeSet[email] = nil
		} else if reference.RefersTo(user) {
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

	// Add all emails based on emails found in course user references.
	// Role based inclusions and exclusions from the course user reference are handled by RefersTo() and Excludes().
	// These calls occurred above (while iterating over the set of users).
	for _, courseReference := range reference.CourseUserReferences {
		for email, _ := range courseReference.Emails {
			_, ok := courseReference.ExcludeEmails[email]
			if ok {
				continue
			}

			_, ok = excludeSet[email]
			if ok {
				continue
			}

			emailSet[email] = nil
		}
	}

	results := make([]string, 0, len(emailSet))

	for email, _ := range emailSet {
		results = append(results, email)
	}

	slices.Sort(results)

	return results
}
