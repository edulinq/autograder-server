package model

import (
	"github.com/edulinq/autograder/internal/timestamp"
)

type TestSubmission struct {
	IgnoreMessages bool   `json:"ignore_messages"`
	SoftError      bool   `json:"soft_error"`
	Stdout         string `json:"stdout"`
	Stderr         string `json:"stderr"`

	GradingInfo *GradingInfo `json:"result"`
}

type SubmissionHistoryItem struct {
	ID               string               `json:"id"`
	ShortID          string               `json:"short-id"`
	CourseID         string               `json:"course-id"`
	AssignmentID     string               `json:"assignment-id"`
	User             string               `json:"user"`
	ProxyUser        string               `json:"proxy-user,omitempty"`
	ProxyTime        *timestamp.Timestamp `json:"proxy-time,omitempty"`
	Message          string               `json:"message"`
	MaxPoints        float64              `json:"max_points"`
	Score            float64              `json:"score"`
	GradingStartTime timestamp.Timestamp  `json:"grading_start_time"`
}

func (this GradingInfo) ToHistoryItem() *SubmissionHistoryItem {
	return &SubmissionHistoryItem{
		ID:               this.ID,
		ShortID:          this.ShortID,
		CourseID:         this.CourseID,
		AssignmentID:     this.AssignmentID,
		User:             this.User,
		ProxyUser:        this.ProxyUser,
		ProxyTime:        this.ProxyStartTime,
		Message:          this.Message,
		MaxPoints:        this.MaxPoints,
		Score:            this.Score,
		GradingStartTime: this.GradingStartTime,
	}
}
