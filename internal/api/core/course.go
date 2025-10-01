package core

import (
	"slices"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

type AssignmentInfo struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	DueDate   *timestamp.Timestamp `json:"due-date,omitempty"`
	MaxPoints float64              `json:"max-points,omitempty"`
}

type CourseInfo struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Assignments map[string]*AssignmentInfo `json:"assignments"`
}

func NewAssignmentInfo(assignment *model.Assignment) *AssignmentInfo {
	return &AssignmentInfo{
		ID:        assignment.ID,
		Name:      assignment.Name,
		DueDate:   assignment.DueDate,
		MaxPoints: assignment.MaxPoints,
	}
}

func NewAssignmentInfos(assignments []*model.Assignment) []*AssignmentInfo {
	result := make([]*AssignmentInfo, 0, len(assignments))
	for _, assignment := range assignments {
		result = append(result, NewAssignmentInfo(assignment))
	}

	return result
}

func NewCourseInfo(course *model.Course) *CourseInfo {
	assignments := make(map[string]*AssignmentInfo, len(course.Assignments))
	for _, assignment := range course.Assignments {
		assignments[assignment.ID] = NewAssignmentInfo(assignment)
	}

	return &CourseInfo{
		ID:          course.ID,
		Name:        course.Name,
		Assignments: assignments,
	}
}

func NewCourseInfos(courses []*model.Course) []*CourseInfo {
	result := make([]*CourseInfo, 0, len(courses))
	for _, course := range courses {
		result = append(result, NewCourseInfo(course))
	}

	return result
}

func NewCourseInfosFromMap(courses map[string]*model.Course) []*CourseInfo {
	ids := make([]string, 0, len(courses))
	for id, _ := range courses {
		ids = append(ids, id)
	}

	slices.Sort(ids)

	result := make([]*CourseInfo, 0, len(courses))
	for _, id := range ids {
		result = append(result, NewCourseInfo(courses[id]))
	}

	return result
}
