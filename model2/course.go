package model2

import (
    "github.com/eriq-augustine/autograder/lms/adapter"
    "github.com/eriq-augustine/autograder/usr"
)

type Course interface {
    GetID() string
    GetName() string
    GetSourceDir() string
    GetLMSAdapter() *adapter.LMSAdapter
    // (LMS ids, assignment ids)
    GetAssignmentLMSIDs() ([]string, []string)
    GetUsers() (map[string]*usr.User, error)
    FullScoringAndUpload(bool) error

    GetSortedAssignments() []Assignment
}
