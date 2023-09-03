package model

type TestSubmission struct {
    IgnoreMessages bool `json:"ignore_messages"`
    Result GradedAssignment `json:"result"`
}
