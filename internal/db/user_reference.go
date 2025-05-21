package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

// Process a list of user inputs in the context of a course.
// See model.CourseUserReferenceInput for the list of acceptable inputs.
// Returns a reference with normalized information, user errors, system error.
// System-level errors immediately return (nil, nil, error).
// User-level errors return (partial reference, aggregated user errors, nil).
func ParseCourseUserReference(course *model.Course, rawReferences []model.CourseUserReferenceInput) (*model.CourseUserReference, error, error) {
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

	var errs error = nil

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
				errs = errors.Join(errs, fmt.Errorf("Unknown course user '%s' in user reference: '%s'.", reference, rawReference))
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
			errs = errors.Join(errs, fmt.Errorf("Unknown course user role '%s' in user reference: '%s'.", reference, rawReference))
			continue
		}

		if exclude {
			courseUserReference.ExcludeCourseUserRoles[reference] = nil
		} else {
			courseUserReference.CourseUserRoles[reference] = nil
		}
	}

	return &courseUserReference, errs, nil
}
