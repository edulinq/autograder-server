package lmssync

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/log"
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
		localID := localAssignment.ID
		localName := localAssignment.Name
		lmsID := localAssignment.LMSID

		assignmentInfo := model.AssignmentInfo{
			ID:   localID,
			Name: localName,
		}

		for i, lmsAssignment := range lmsAssignments {
			matchIndex := -1

			if lmsID != "" {
				// Exact ID match.
				if lmsID == lmsAssignment.ID {
					matchIndex = i
				}
			} else {
				// Name match.
				if (localName != "") && approximateNameMatch(localName, lmsAssignment.Name) {
					matchIndex = i
				}
			}

			if matchIndex != -1 {
				_, exists := matches[localID]
				if exists {
					delete(matches, localID)
					result.AmbiguousMatches = append(result.AmbiguousMatches, assignmentInfo)
					break
				}

				matches[localID] = matchIndex
			}
		}

		_, exists := matches[localID]
		if !exists {
			result.NonMatchedAssignments = append(result.NonMatchedAssignments, assignmentInfo)
		}
	}

	for localID, lmsIndex := range matches {
		localAssignment := localAssignments[localID]
		localName := localAssignment.GetDisplayName()

		// Check all matched assignments for a matching late days assignment.
		lateDaysAssignment := matchLateDaysAssignment(localAssignment, lmsAssignments)

		assignmentInfo := model.AssignmentInfo{
			ID:   localID,
			Name: localName,
		}

		if lateDaysAssignment != nil {
			assignmentInfo.LateDaysLMSID = lateDaysAssignment.ID
			assignmentInfo.LateDaysLMSName = lateDaysAssignment.Name
		}

		changed := mergeAssignment(localAssignment, lmsAssignments[lmsIndex], lateDaysAssignment)
		if changed {
			result.SyncedAssignments = append(result.SyncedAssignments, assignmentInfo)
		} else {
			result.UnchangedAssignments = append(result.UnchangedAssignments, assignmentInfo)
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

// Matching late days is more strict.
// First, look for a strict ID match.
// Then (only after all assignments have been checked), look for an approximate name match.
// If there is ambiguity on the name match, don't match.
func matchLateDaysAssignment(localAssignment *model.Assignment, lmsAssignments []*lmstypes.Assignment) *lmstypes.Assignment {
	lateLMSID := localAssignment.LatePolicy.LateDaysLMSID
	lateLMSName := localAssignment.LatePolicy.LateDaysLMSName

	if (lateLMSID == "") && (lateLMSName == "") {
		return nil
	}

	for _, lmsAssignment := range lmsAssignments {
		if (lateLMSID != "") && (lateLMSID == lmsAssignment.ID) {
			return lmsAssignment
		}
	}

	var match *lmstypes.Assignment = nil
	for _, lmsAssignment := range lmsAssignments {
		if (lateLMSName != "") && approximateNameMatch(lateLMSName, lmsAssignment.Name) {
			if match != nil {
				log.Warn("Ambiguous late days match for assignment.",
					localAssignment, log.NewAttr("match-1", match.Name), log.NewAttr("match-2", lmsAssignment.Name))

				return nil
			}

			match = lmsAssignment
		}
	}

	return match
}

func approximateNameMatch(a string, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	return strings.EqualFold(a, b)
}

func mergeAssignment(localAssignment *model.Assignment, lmsAssignment *lmstypes.Assignment, lateLMSAssignment *lmstypes.Assignment) bool {
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

	if lateLMSAssignment != nil {
		if localAssignment.LatePolicy.LateDaysLMSID == "" {
			localAssignment.LatePolicy.LateDaysLMSID = lateLMSAssignment.ID
			changed = true
		}

		if (localAssignment.LatePolicy.LateDaysLMSName == "") && (lateLMSAssignment.Name != "") {
			localAssignment.LatePolicy.LateDaysLMSName = lateLMSAssignment.Name
			changed = true
		}
	}

	return changed
}
