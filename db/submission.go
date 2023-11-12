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
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    course, ok := rawCourse.(*types.Course);
    if (!ok) {
        return fmt.Errorf("Course '%v' is not a db course.", rawCourse);
    }

    return backend.SaveSubmissions(course, submissions);
}

func SaveSubmission(rawAssignment model.Assignment, submission *artifact.GradingResult) error {
    if (backend == nil) {
        return fmt.Errorf("Database has not been opened.");
    }

    return SaveSubmissions(rawAssignment.GetCourse(), []*artifact.GradingResult{submission});
}

func GetNextSubmissionID(rawAssignment model.Assignment, email string) (string, error) {
    if (backend == nil) {
        return "", fmt.Errorf("Database has not been opened.");
    }

    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return "", fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetNextSubmissionID(assignment, email);
}

func GetSubmissionHistory(rawAssignment model.Assignment, email string) ([]*artifact.SubmissionHistoryItem, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetSubmissionHistory(assignment, email);
}

func GetSubmissionResult(rawAssignment model.Assignment, email string, submissionID string) (*artifact.GradingInfo, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    shortSubmissionID := common.GetShortSubmissionID(submissionID);
    return backend.GetSubmissionResult(assignment, email, shortSubmissionID);
}

func GetScoringInfos(rawAssignment model.Assignment, filterRole usr.UserRole) (map[string]*artifact.ScoringInfo, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetScoringInfos(assignment, filterRole);
}

func GetRecentSubmissions(rawAssignment model.Assignment, filterRole usr.UserRole) (map[string]*artifact.GradingInfo, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetRecentSubmissions(assignment, filterRole);
}

func GetRecentSubmissionSurvey(rawAssignment model.Assignment, filterRole usr.UserRole) (map[string]*artifact.SubmissionHistoryItem, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetRecentSubmissionSurvey(assignment, filterRole);
}

func GetSubmissionContents(rawAssignment model.Assignment, email string, submissionID string) (*artifact.GradingResult, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    shortSubmissionID := common.GetShortSubmissionID(submissionID);
    return backend.GetSubmissionContents(assignment, email, shortSubmissionID);
}

func GetRecentSubmissionContents(rawAssignment model.Assignment, filterRole usr.UserRole) (map[string]*artifact.GradingResult, error) {
    if (backend == nil) {
        return nil, fmt.Errorf("Database has not been opened.");
    }

    assignment, ok := rawAssignment.(*types.Assignment);
    if (!ok) {
        return nil, fmt.Errorf("Assignment '%v' is not a db assignment.", rawAssignment);
    }

    return backend.GetRecentSubmissionContents(assignment, filterRole);
}
