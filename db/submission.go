package db

import (
    "fmt"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/model"
)

func SaveSubmissions(course *model.Course, submissions []*model.GradingResult) error {
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    return backend.SaveSubmissions(course, submissions);
}

func SaveSubmission(assignment *model.Assignment, submission *model.GradingResult) error {
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    return SaveSubmissions(assignment.GetCourse(), []*model.GradingResult{submission});
}

func GetNextSubmissionID(assignment *model.Assignment, email string) (string, error) {
    if (backend == nil) {
        return "", fmt.Errorf("Database has not been opened.");
    }

    return backend.GetNextSubmissionID(assignment, email);
}

func GetSubmissionHistory(assignment *model.Assignment, email string) ([]*model.SubmissionHistoryItem, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return backend.GetSubmissionHistory(assignment, email);
}

func GetSubmissionResult(assignment *model.Assignment, email string, submissionID string) (*model.GradingInfo, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    shortSubmissionID := common.GetShortSubmissionID(submissionID);
    return backend.GetSubmissionResult(assignment, email, shortSubmissionID);
}

func GetScoringInfos(assignment *model.Assignment, filterRole model.UserRole) (map[string]*model.ScoringInfo, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return backend.GetScoringInfos(assignment, filterRole);
}

func GetRecentSubmissions(assignment *model.Assignment, filterRole model.UserRole) (map[string]*model.GradingInfo, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return backend.GetRecentSubmissions(assignment, filterRole);
}

func GetRecentSubmissionSurvey(assignment *model.Assignment, filterRole model.UserRole) (map[string]*model.SubmissionHistoryItem, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return backend.GetRecentSubmissionSurvey(assignment, filterRole);
}

func GetSubmissionContents(assignment *model.Assignment, email string, submissionID string) (*model.GradingResult, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    shortSubmissionID := common.GetShortSubmissionID(submissionID);
    return backend.GetSubmissionContents(assignment, email, shortSubmissionID);
}

func GetRecentSubmissionContents(assignment *model.Assignment, filterRole model.UserRole) (map[string]*model.GradingResult, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    return backend.GetRecentSubmissionContents(assignment, filterRole);
}

func RemoveSubmission(assignment *model.Assignment, email string, submissionID string) (bool, error) {
    if (backend == nil) {
        return false, fmt.Errorf("Database has not been opened.");
    }

    shortSubmissionID := common.GetShortSubmissionID(submissionID);
    return backend.RemoveSubmission(assignment, email, shortSubmissionID);
}
