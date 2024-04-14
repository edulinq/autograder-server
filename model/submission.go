package model

import (
    "github.com/edulinq/autograder/common"
)

type TestSubmission struct {
    IgnoreMessages bool `json:"ignore_messages"`
    GradingInfo *GradingInfo `json:"result"`
    Error string `json:"error"`
}

type SubmissionHistoryItem struct {
    ID string `json:"id"`
    ShortID string `json:"short-id"`
    CourseID string `json:"course-id"`
    AssignmentID string `json:"assignment-id"`
    User string `json:"user"`
    Message string `json:"message"`
    MaxPoints float64 `json:"max_points"`
    Score float64 `json:"score"`
    GradingStartTime common.Timestamp `json:"grading_start_time"`
}

func (this GradingInfo) ToHistoryItem() *SubmissionHistoryItem {
    return &SubmissionHistoryItem{
        ID: this.ID,
        ShortID: this.ShortID,
        CourseID: this.CourseID,
        AssignmentID: this.AssignmentID,
        User: this.User,
        Message: this.Message,
        MaxPoints: this.MaxPoints,
        Score: this.Score,
        GradingStartTime: this.GradingStartTime,
    };
}
