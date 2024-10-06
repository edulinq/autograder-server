package upsert

import (
	"github.com/edulinq/autograder/internal/procedures/courses"
)

type UpsertResponse struct {
	Results []courses.CourseUpsertResult `json:"results"`
}
