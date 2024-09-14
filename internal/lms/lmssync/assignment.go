package lmssync

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func syncAssignments(course *model.Course, dryRun bool) (*model.AssignmentSyncResult, error) {
	result := model.NewAssignmentSyncResult()

	adapter := course.GetLMSAdapter()
	if (adapter == nil) || (!adapter.SyncAssignments) {
		return result, nil
	}

	lmsAssignments, err := lms.FetchAssignments(course)
	if err != nil {
		return nil, fmt.Errorf("Failed to get assignments: '%w'.", err)
	}

	localAssignments := course.GetAssignments()

	// Match local assignments to LMS assignments.
	matches := make(map[string]int)
	for _, localAssignment := range localAssignments {
		localID := localAssignment.GetID()
		localName := localAssignment.GetName()
		lmsID := localAssignment.GetLMSID()

		for i, lmsAssignment := range lmsAssignments {
			matchIndex := -1

			if lmsID != "" {
				// Exact ID match.
				if lmsID == lmsAssignment.ID {
					matchIndex = i
				}
			} else {
				// Name match.
				if (localName != "") && strings.EqualFold(localName, lmsAssignment.Name) {
					matchIndex = i
				}
			}

			if matchIndex != -1 {
				_, exists := matches[localID]
				if exists {
					delete(matches, localID)
					result.AmbiguousMatches = append(result.AmbiguousMatches, model.AssignmentInfo{localID, localName})
					break
				}

				matches[localID] = matchIndex
			}
		}

		_, exists := matches[localID]
		if !exists {
			result.NonMatchedAssignments = append(result.NonMatchedAssignments, model.AssignmentInfo{localID, localName})
		}
	}

	for localID, lmsIndex := range matches {
		localName := localAssignments[localID].GetName()
		changed := mergeAssignment(localAssignments[localID], lmsAssignments[lmsIndex])
		if changed {
			result.SyncedAssignments = append(result.SyncedAssignments, model.AssignmentInfo{localID, localName})
		} else {
			result.UnchangedAssignments = append(result.UnchangedAssignments, model.AssignmentInfo{localID, localName})
		}
	}

	if !dryRun {
		err = db.SaveCourse(course)
		if err != nil {
			return nil, fmt.Errorf("Failed to save course: '%w'.", err)
		}
	}

	return result, nil
}

func mergeAssignment(localAssignment *model.Assignment, lmsAssignment *lmstypes.Assignment) bool {
	changed := false

	if localAssignment.LMSID == "" {
		localAssignment.LMSID = lmsAssignment.ID
		changed = true
	}

	if (localAssignment.Name == "") && (lmsAssignment.Name != "") {
		localAssignment.Name = lmsAssignment.Name
		changed = true
	}

	if (localAssignment.DueDate == nil) && (lmsAssignment.DueDate != nil) {
		localAssignment.DueDate = lmsAssignment.DueDate
		changed = true
	}

	if util.IsZero(localAssignment.MaxPoints) && !util.IsZero(lmsAssignment.MaxPoints) {
		localAssignment.MaxPoints = lmsAssignment.MaxPoints
		changed = true
	}

	return changed
}
