package core

import (
	"github.com/edulinq/autograder/internal/model"
)

// An API-safe representation of an assignment.
type AssignmentInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewAssignmentInfo(assignment *model.Assignment) *AssignmentInfo {
	return &AssignmentInfo{
		ID:   assignment.GetID(),
		Name: assignment.GetName(),
	}
}

// Assignments are output in the same order they are input.
func NewAssignmentInfos(assignments []*model.Assignment) []*AssignmentInfo {
	result := make([]*AssignmentInfo, 0, len(assignments))
	for _, assignment := range assignments {
		result = append(result, NewAssignmentInfo(assignment))
	}

	return result
}
