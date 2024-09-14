package logs

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

func validateQuery(query *log.ParsedLogQuery) error {
	if query == nil {
		return fmt.Errorf("Cannot validate a nil query.")
	}

	var err error = nil
	var course *model.Course = nil

	if query.CourseID != "" {
		course, err = validateCourse(query.CourseID)
		if err != nil {
			return err
		}

		if course != nil {
			query.CourseID = course.GetID()
		}
	}

	if query.AssignmentID != "" {
		assignment, err := validateAssignment(query.AssignmentID, course)
		if err != nil {
			return err
		}

		if assignment != nil {
			query.AssignmentID = assignment.GetID()
		}
	}

	if query.UserEmail != "" {
		user, err := validateUser(query.UserEmail, course)
		if err != nil {
			return err
		}

		if user != nil {
			query.UserEmail = user.Email
		}
	}

	return nil
}

func validateCourse(courseID string) (*model.Course, error) {
	course, err := db.GetCourse(courseID)
	if err != nil {
		return nil, err
	}

	if course == nil {
		return nil, fmt.Errorf("Could not find course with ID '%s'.", courseID)
	}

	return course, nil
}

func validateAssignment(assignmentID string, course *model.Course) (*model.Assignment, error) {
	if course == nil {
		return nil, fmt.Errorf("When an assignment is provided, a course must also be provided.")
	}

	assignment := course.GetAssignment(assignmentID)
	if assignment == nil {
		return nil, fmt.Errorf("Could not find assignment with ID '%s' in course '%s'.", assignmentID, course.GetID())
	}

	return assignment, nil
}

func validateUser(userEmail string, course *model.Course) (*model.ServerUser, error) {
	user, err := db.GetServerUser(userEmail)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Could not find user with email '%s'.", userEmail)
	}

	if (course != nil) && (!user.IsEnrolled(course.GetID())) {
		return nil, fmt.Errorf("A course ('%s') and user ('%s') was provided, but the user is not enrolled in that course.", course.GetID(), user.Email)
	}

	return user, nil
}
