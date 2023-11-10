package db

import (
    "fmt"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
)

func SaveSubmissions(rawCourse model.Course, submissions []*artifact.GradingResult) error {
    course, ok := rawCourse.(*types.Course);
    if (!ok) {
        return fmt.Errorf("Course '%v' is not a db course.", rawCourse);
    }

    return backend.SaveSubmissions(course, submissions);
}

func SaveSubmission(rawCourse model.Course, submission *artifact.GradingResult) error {
    return SaveSubmissions(rawCourse, []*artifact.GradingResult{submission});
}

func GetNextSubmissionID(rawAssignment model.Assignment, email string) (string, error) {
    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return "", fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetNextSubmissionID(assignment, email);
}

func GetSubmissionHistory(rawAssignment model.Assignment, email string) ([]*artifact.SubmissionHistoryItem, error) {
    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetSubmissionHistory(assignment, email);
}

func GetSubmissionResult(rawAssignment model.Assignment, email string, submissionID string) (*artifact.GradedAssignment, error) {
    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    shortSubmissionID := common.GetShortSubmissionID(submissionID);
    return backend.GetSubmissionResult(assignment, email, shortSubmissionID);
}

func GetScoringInfos(rawAssignment model.Assignment, onlyRole usr.UserRole) (map[string]*artifact.ScoringInfo, error) {
    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetScoringInfos(assignment, onlyRole);
}

func GetRecentSubmissions(rawAssignment model.Assignment, onlyRole usr.UserRole) (map[string]*artifact.GradedAssignment, error) {
    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetRecentSubmissions(assignment, onlyRole);
}

func GetRecentSubmissionSurvey(rawAssignment model.Assignment, onlyRole usr.UserRole) (map[string]*artifact.SubmissionHistoryItem, error) {
    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetRecentSubmissionSurvey(assignment, onlyRole);
}
