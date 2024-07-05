package model

type AssignmentSyncResult struct {
	SyncedAssignments     []AssignmentInfo `json:"synced-assignments"`
	AmbiguousMatches      []AssignmentInfo `json:"ambiguous-matches"`
	NonMatchedAssignments []AssignmentInfo `json:"non-matched-assignments"`
	UnchangedAssignments  []AssignmentInfo `json:"unchanged-assignments"`
}

type AssignmentInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAssignmentSyncResult() *AssignmentSyncResult {
	return &AssignmentSyncResult{
		SyncedAssignments:     make([]AssignmentInfo, 0),
		AmbiguousMatches:      make([]AssignmentInfo, 0),
		NonMatchedAssignments: make([]AssignmentInfo, 0),
		UnchangedAssignments:  make([]AssignmentInfo, 0),
	}
}
