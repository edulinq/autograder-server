package lmstypes

import (
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

type User struct {
	ID    string
	Name  string
	Email string
	Role  model.CourseUserRole
}

type SubmissionScore struct {
	UserID   string
	Score    float64
	Time     *timestamp.Timestamp
	Comments []*SubmissionComment
}

type SubmissionComment struct {
	ID     string
	Author string
	Text   string
	Time   string
}

type Assignment struct {
	ID          string
	Name        string
	LMSCourseID string
	DueDate     *timestamp.Timestamp
	MaxPoints   float64
}

func (this *User) ToRawUserData(courseID string) *model.RawUserData {
	data := &model.RawUserData{
		Email:       this.Email,
		Name:        this.Name,
		Course:      courseID,
		CourseRole:  this.Role.String(),
		CourseLMSID: this.ID,
	}

	return data
}
