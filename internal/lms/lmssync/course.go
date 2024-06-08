package lmssync

import (
	"github.com/edulinq/autograder/internal/model"
)

// Sync all available aspects of the course with their LMS.
// Will return nil (with no error) if the course has no LMS.
func SyncLMS(course *model.Course, dryRun bool, sendEmails bool) (*model.LMSSyncResult, error) {
	if !course.HasLMSAdapter() {
		return nil, nil
	}

	userSync, err := SyncAllLMSUsers(course, dryRun, sendEmails)
	if err != nil {
		return nil, err
	}

	assignmentSync, err := syncAssignments(course, dryRun)
	if err != nil {
		return nil, err
	}

	result := &model.LMSSyncResult{
		UserSync:       userSync,
		AssignmentSync: assignmentSync,
	}

	return result, nil
}
