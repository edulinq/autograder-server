package lms

import (
    "time"

    "github.com/eriq-augustine/autograder/usr"
)

// An interface for adapters to work closely with courses.
type SourceCourse interface {
    GetSourceDir() string
    GetUsers() (map[string]*usr.User, error)
}

type User struct {
    ID string
    Name string
    Email string
    Role usr.UserRole
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
