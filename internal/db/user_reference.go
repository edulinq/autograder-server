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
	serverUserReference := model.ServerUserReference{
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		ServerUserRoles:        make(map[string]model.ServerUserRole, 0),
		ExcludeServerUserRoles: make(map[string]model.ServerUserRole, 0),
		CourseUserReferences:   make(map[string]*model.CourseUserReference, 0),
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
				serverUserReference.ExcludeServerUserRoles = model.GetCommonServerUserRoleStrings()
			} else {
				serverUserReference.ServerUserRoles = model.GetCommonServerUserRoleStrings()
			}

			continue
		}

		parts := strings.Split(reference, model.USER_REFERENCE_DELIM)
		if len(parts) == 1 {
			// User reference must be a server role.
			commonServerRoles := model.GetCommonServerUserRoleStrings()
			serverUserRole, ok := commonServerRoles[reference]
			if !ok {
				errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'. Unknown server user role '%s'.", i, rawReference, reference))
				continue
			}

			if exclude {
				serverUserReference.ExcludeServerUserRoles[reference] = serverUserRole
			} else {
				serverUserReference.ServerUserRoles[reference] = serverUserRole
			}
		} else if len(parts) == 2 {
			// User reference must be <course-id>::<course-role>.
			// If a '*' is present, target all courses or course roles respectively.
			courseID := strings.ToLower(strings.TrimSpace(parts[0]))
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
					errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'. Unknown course '%s'.", i, rawReference, courseID))
					continue
				}

				courses[course.GetID()] = course
			}

			courseRoles := make(map[string]model.CourseUserRole, 0)

			if courseRoleString == "*" {
				// Target all course roles.
				courseRoles = model.GetCommonCourseUserRoleStrings()
			} else {
				// Target a specific course role.
				commonCourseRoles := model.GetCommonCourseUserRoleStrings()
				courseRole, ok := commonCourseRoles[courseRoleString]
				if !ok {
					errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'. Unknown course user role '%s'.", i, rawReference, courseRoleString))
					continue
				}

				courseRoles[courseRoleString] = courseRole
			}

			for _, course := range courses {
				courseUserReference := createCourseUserReference(course, courseRoles, exclude)
				serverUserReference.AddCourseUserReference(courseUserReference)
			}
		} else {
			errs = errors.Join(errs, fmt.Errorf("Invalid user reference format: '%s'.", rawReference))
			continue
		}
	}

	return &serverUserReference, errs
}

// Process a list of user inputs in the context of a course.
// See model.ServerUserReferenceInput for the list of acceptable inputs.
// Returns a reference with normalized information and error.
// System-level errors immediately return (nil, error).
// User-level errors return (partial reference, aggregated user errors).
func ParseCourseUserReference(course *model.Course, rawReferences []model.CourseUserReferenceInput) (*model.CourseUserReference, error) {
	courseUserReference := model.CourseUserReference{
		Course:                 course,
		Emails:                 make(map[string]any, 0),
		ExcludeEmails:          make(map[string]any, 0),
		CourseUserRoles:        make(map[string]model.CourseUserRole, 0),
		ExcludeCourseUserRoles: make(map[string]model.CourseUserRole, 0),
	}

	var errs error = nil

	commonCourseRoles := model.GetCommonCourseUserRoleStrings()
	courseUsers, err := GetCourseUsers(course)
	if err != nil {
		return nil, fmt.Errorf("Failed to get courses: '%w'.", err)
	}

	for i, rawReference := range rawReferences {
		reference := strings.ToLower(strings.TrimSpace(string(rawReference)))

		exclude := false
		if strings.HasPrefix(reference, "-") {
			exclude = true

			reference = strings.TrimPrefix(reference, "-")
		}

		if strings.Contains(reference, "@") {
			_, ok := courseUsers[reference]
			if !ok {
				errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'. Unknown course user '%s'.", i, rawReference, reference))
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
		courseRole, ok := commonCourseRoles[reference]
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("Unknown user reference %d: '%s'. Unknown course user role '%s'.", i, rawReference, reference))
			continue
		}

		if exclude {
			courseUserReference.ExcludeCourseUserRoles[reference] = courseRole
		} else {
			courseUserReference.CourseUserRoles[reference] = courseRole
		}
	}

	return &courseUserReference, errs
}

func createCourseUserReference(course *model.Course, courseRoles map[string]model.CourseUserRole, exclude bool) *model.CourseUserReference {
	courseUserRoles := make(map[string]model.CourseUserRole, 0)
	excludeCourseUserRoles := make(map[string]model.CourseUserRole, 0)

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
