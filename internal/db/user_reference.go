package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

// Process a list of user inputs and return a normalized reference and error.
// See model.ServerUserReferenceInput for the list of acceptable inputs.
// System-level errors immediately return (nil, error).
// User-level errors return (partial reference, aggregated user errors).
func ParseServerUserReference(rawReferences []model.ServerUserReferenceInput) (*model.ServerUserReference, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	serverUserReference := model.ServerUserReference{
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		ServerUserRoles:        make(map[string]any, 0),
		ExcludeServerUserRoles: make(map[string]any, 0),
		CourseUserReferences:   make(map[string]*model.CourseUserReference, 0),
	}

	var errs error = nil
	var err error = nil

	commonServerRoles := model.GetCommonServerUserRoleStrings()
	commonCourseRoles := model.GetCommonCourseUserRoleStrings()

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
			if exclude {
				serverUserReference.ExcludeEmails[reference] = nil
			} else {
				serverUserReference.Emails[reference] = nil
			}

			continue
		}

		if reference == "*" {
			if exclude {
				serverUserReference.ExcludeServerUserRoles = commonServerRoles
			} else {
				serverUserReference.ServerUserRoles = commonServerRoles
			}

			continue
		}

		parts := strings.Split(reference, model.USER_REFERENCE_DELIM)
		if len(parts) == 1 {
			// User reference must be a server role.
			_, ok := commonServerRoles[reference]
			if !ok {
				errs = errors.Join(errs, fmt.Errorf("Unknown server user role '%s' in user reference: '%s'.", reference, rawReference))
				continue
			}

			if exclude {
				serverUserReference.ExcludeServerUserRoles[reference] = nil
			} else {
				serverUserReference.ServerUserRoles[reference] = nil
			}
		} else if len(parts) == 2 {
			// User reference must be <course-id>::<course-role>.
			// If a '*' is present, target all courses or course roles respectively.
			courseID := strings.TrimSpace(parts[0])
			courseRoleString := strings.TrimSpace(parts[1])

			courses := make(map[string]*model.Course, 0)
			if courseID == "*" {
				// Target all courses.
				courses, err = GetCourses()
				if err != nil {
					return nil, fmt.Errorf("Failed to get courses: '%w'.", err)
				}
			} else {
				// Target a specific course.
				course, err := GetCourse(courseID)
				if err != nil {
					return nil, fmt.Errorf("Failed to get course '%s': '%w'.", courseID, err)
				}

				if course == nil {
					errs = errors.Join(errs, fmt.Errorf("Unknown course '%s' in user reference: '%s'.", courseID, rawReference))
					continue
				}

				courses[course.GetID()] = course
			}

			courseRoles := make(map[string]any, 0)
			if courseRoleString == "*" {
				// Target all course roles.
				courseRoles = commonCourseRoles
			} else {
				// Target a specific course role.
				_, ok := commonCourseRoles[courseRoleString]
				if !ok {
					errs = errors.Join(errs, fmt.Errorf("Unknown course user role '%s' in user reference: '%s'.", courseRoleString, rawReference))
					continue
				}

				courseRoles[courseRoleString] = nil
			}

			for _, course := range courses {
				courseUserReference := createCourseUserReference(course, courseRoles, exclude)
				serverUserReference.AddCourseUserReference(courseUserReference)
			}
		} else {
			errs = errors.Join(errs, fmt.Errorf("Invalid format in user reference: '%s'.", rawReference))
			continue
		}
	}

	return &serverUserReference, errs
}

// Process a list of user inputs in the context of a course.
// See model.CourseUserReferenceInput for the list of acceptable inputs.
// Returns a reference with normalized information and error.
// System-level errors immediately return (nil, error).
// User-level errors return (partial reference, aggregated user errors).
func ParseCourseUserReference(course *model.Course, rawReferences []model.CourseUserReferenceInput) (*model.CourseUserReference, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
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
		return nil, fmt.Errorf("Failed to get courses: '%w'.", err)
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

	return &courseUserReference, errs
}

func createCourseUserReference(course *model.Course, courseRoles map[string]any, exclude bool) *model.CourseUserReference {
	courseUserRoles := make(map[string]any, 0)
	excludeCourseUserRoles := make(map[string]any, 0)

	if exclude {
		excludeCourseUserRoles = courseRoles
	} else {
		courseUserRoles = courseRoles
	}

	return &model.CourseUserReference{
		Course:                 course,
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		CourseUserRoles:        courseUserRoles,
		ExcludeCourseUserRoles: excludeCourseUserRoles,
	}
}
