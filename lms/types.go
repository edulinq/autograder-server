package lms

import (
    "time"
)

type User struct {
    ID string
    Name string
    Email string
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
    CourseID string
    DueDate *time.Time
    MaxPoints float64
}
