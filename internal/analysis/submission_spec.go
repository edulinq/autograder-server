package analysis

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

// Resolve submission specs into full submission IDs.
// A submission spec may take the following forms:
// - "<course id>::<assignment id>::<user email>::<submission short id>" - This is the same as a full submission id.
// - "<course id>::<assignment id>::<user email>" - The user's most recent submission for this assignment.
// - "<course id>::<assignment id>" - The most recent submission for all students (not just users).
// Returns: (resolved full submission ids, all seen courses, user errors, system errors)
func ResolveSubmissionSpecs(submissionSpecs []string) ([]string, []string, error, error) {
	submissionIDs := make(map[string]bool)
	seenCourses := make(map[string]bool)

	var systemErrors error = nil
	var userErrors error = nil

	for i, submissionSpec := range submissionSpecs {
		courseID, assignmentID, userEmail, submissionShortID, err := splitSubmissionSpec(submissionSpec)
		if err != nil {
			userErrors = errors.Join(userErrors, fmt.Errorf("Submission spec at index %d had an error: '%w'.", i, err))
			continue
		}

		course, err := db.GetCourse(courseID)
		if err != nil {
			systemErrors = errors.Join(systemErrors, fmt.Errorf("Failed to fetch course (%s): '%w'.", courseID, err))
			continue
		}

		if course == nil {
			userErrors = errors.Join(userErrors, fmt.Errorf("Course not found: %s.", courseID))
			continue
		}

		assignment := course.GetAssignment(assignmentID)
		if assignment == nil {
			userErrors = errors.Join(userErrors, fmt.Errorf("Assignment not found %s.%s.", courseID, assignmentID))
			continue
		}

		if userEmail == "" {
			// Most recent submissions for entire course.
			submissions, err := db.GetRecentSubmissions(assignment, model.CourseRoleStudent)
			if err != nil {
				systemErrors = errors.Join(systemErrors, fmt.Errorf("Failed to fetch submissions (%s): '%w'.", submissionSpec, err))
				continue
			}

			for _, submission := range submissions {
				if submission == nil {
					continue
				}

				submissionIDs[submission.ID] = true
				seenCourses[courseID] = true
			}
		} else {
			// Most recent submission for user or single submission.
			// The DB uses the same function for both of these (with an empty submission short ID for the former).
			submission, err := db.GetSubmissionResult(assignment, userEmail, submissionShortID)
			if err != nil {
				systemErrors = errors.Join(systemErrors, fmt.Errorf("Failed to fetch submission (%s): '%w'.", submissionSpec, err))
				continue
			}

			if submission == nil {
				userErrors = errors.Join(userErrors, fmt.Errorf("Could not find submission %s.", submissionSpec))
				continue
			}

			submissionIDs[submission.ID] = true
			seenCourses[courseID] = true
		}
	}

	ids := make([]string, 0, len(submissionIDs))
	for id, _ := range submissionIDs {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	courses := make([]string, 0, len(seenCourses))
	for course, _ := range seenCourses {
		courses = append(courses, course)
	}
	slices.Sort(courses)

	if (userErrors != nil) || (systemErrors != nil) {
		return nil, nil, userErrors, systemErrors
	}

	return ids, courses, nil, nil
}

func splitSubmissionSpec(submissionSpec string) (string, string, string, string, error) {
	parts := strings.Split(submissionSpec, common.SUBMISSION_ID_DELIM)

	if len(parts) > 4 {
		return "", "", "", "", fmt.Errorf("Submission spec has too many components %d. Max is 4.", len(parts))
	}

	if len(parts) < 2 {
		return "", "", "", "", fmt.Errorf("Submission spec has too few components %d. Min is 2.", len(parts))
	}

	for j, part := range parts {
		if part == "" {
			return "", "", "", "", fmt.Errorf("Submission spec has an empty component at index %d.", j)
		}
	}

	rawCourseID := parts[0]
	rawAssignmentID := parts[1]

	userEmail := ""
	if len(parts) >= 3 {
		userEmail = parts[2]
	}

	submissionID := ""
	if len(parts) >= 4 {
		submissionID = parts[3]

		if !regexp.MustCompile(`^\d+$`).MatchString(submissionID) {
			return "", "", "", "", fmt.Errorf("Submission short ID is invalid (%s).", submissionID)
		}
	}

	courseID, err := common.ValidateID(rawCourseID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("Course ID is invalid (%s): '%w'.", rawCourseID, err)
	}

	assignmentID, err := common.ValidateID(rawAssignmentID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("Assignment ID is invalid (%s): '%w'.", rawAssignmentID, err)
	}

	return courseID, assignmentID, userEmail, submissionID, err
}
