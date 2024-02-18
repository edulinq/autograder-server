package lmstypes

import (
    "time"

    "github.com/edulinq/autograder/model"
)

type User struct {
    ID string
    Name string
    Email string
    Role model.UserRole
}

type SubmissionScore struct {
    UserID string
    Score float64
    Time time.Time
    Comments []*SubmissionComment
}

type SubmissionComment struct {
    ID string
    Author string
    Text string
    Time string
}

type Assignment struct {
    ID string
    Name string
    LMSCourseID string
    DueDate *time.Time
    MaxPoints float64
}
