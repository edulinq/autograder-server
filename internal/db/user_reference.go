package db

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

// Process a list of user inputs in the context of a course.
// See model.CourseUserReferenceInput for the list of acceptable inputs.
// Returns a reference, user errors, and a system error.
// User-level errors return (partial reference, user errors, nil).
// A system-level error returns (nil, nil, error).
func ParseCourseUserReference(course *model.Course, rawReferences []model.CourseUserReferenceInput) (*model.CourseUserReference, map[string]error, error) {
	if backend == nil {
		return nil, nil, fmt.Errorf("Database has not been opened.")
	}

	courseUserReference := model.CourseUserReference{
		Course:                 course,
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		CourseUserRoles:        make(map[string]any, 0),
		ExcludeCourseUserRoles: make(map[string]any, 0),
	}

	workErrors := make(map[string]error, 0)

	commonCourseRoles := model.GetCommonCourseUserRoleStrings()

	courseUsers, err := GetCourseUsers(course)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get courses: '%w'.", err)
	}

	for _, rawReference := range rawReferences {
		reference := strings.ToLower(strings.TrimSpace(string(rawReference)))

		exclude := false
		if strings.HasPrefix(reference, "-") {
			exclude = true

			reference = strings.TrimPrefix(reference, "-")
			reference = strings.TrimSpace(reference)
		}

		if reference == "" {
			continue
		}

		if strings.Contains(reference, "@") {
			_, ok := courseUsers[reference]
			if !ok {
				workErrors[string(rawReference)] = fmt.Errorf("Unknown course user: '%s'.", reference)
				continue
			}

			if exclude {
				courseUserReference.ExcludeEmails[reference] = nil
			} else {
				courseUserReference.Emails[reference] = nil
			}

			continue
		}

		if reference == "*" {
			if exclude {
				courseUserReference.ExcludeCourseUserRoles = commonCourseRoles
			} else {
				courseUserReference.CourseUserRoles = commonCourseRoles
			}

			continue
		}

		// Target a specific course role.
		_, ok := commonCourseRoles[reference]
		if !ok {
			workErrors[string(rawReference)] = fmt.Errorf("Unknown course role: '%s'.", reference)
			continue
		}

		if exclude {
			courseUserReference.ExcludeCourseUserRoles[reference] = nil
		} else {
			courseUserReference.CourseUserRoles[reference] = nil
		}
	}

	return &courseUserReference, workErrors, nil
}
