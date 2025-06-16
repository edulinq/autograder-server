package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/model"
)

func SaveSubmissions(course *model.Course, submissions []*model.GradingResult) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.SaveSubmissions(course, submissions)
}

func SaveSubmission(assignment *model.Assignment, submission *model.GradingResult) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return SaveSubmissions(assignment.GetCourse(), []*model.GradingResult{submission})
}

func GetNextSubmissionID(assignment *model.Assignment, email string) (string, error) {
	if backend == nil {
		return "", fmt.Errorf("Database has not been opened.")
	}

	return backend.GetNextSubmissionID(assignment, email)
}

func GetPreviousSubmissionID(assignment *model.Assignment, email string, submissionID string) (string, error) {
	if backend == nil {
		return "", fmt.Errorf("Database has not been opened.")
	}

	shortSubmissionID := common.GetShortSubmissionID(submissionID)
	return backend.GetPreviousSubmissionID(assignment, email, shortSubmissionID)
}

func GetSubmissionHistory(assignment *model.Assignment, email string) ([]*model.SubmissionHistoryItem, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetSubmissionHistory(assignment, email)
}

func GetSubmissionResult(assignment *model.Assignment, email string, submissionID string) (*model.GradingInfo, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	shortSubmissionID := common.GetShortSubmissionID(submissionID)
	return backend.GetSubmissionResult(assignment, email, shortSubmissionID)
}

// Get only non-nil scoring infos.
func GetExistingScoringInfos(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.ScoringInfo, error) {
	rawInfo, err := GetScoringInfos(assignment, reference)
	if err != nil {
		return nil, err
	}

	info := make(map[string]*model.ScoringInfo, len(rawInfo))
	for key, value := range rawInfo {
		if value != nil {
			info[key] = value
		}
	}

	return info, nil
}

func GetScoringInfos(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.ScoringInfo, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetScoringInfos(assignment, reference)
}

func GetRecentSubmissions(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.GradingInfo, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetRecentSubmissions(assignment, reference)
}

func GetRecentSubmissionSurvey(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.SubmissionHistoryItem, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetRecentSubmissionSurvey(assignment, reference)
}

func GetSubmissionContents(assignment *model.Assignment, email string, submissionID string) (*model.GradingResult, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	shortSubmissionID := common.GetShortSubmissionID(submissionID)
	return backend.GetSubmissionContents(assignment, email, shortSubmissionID)
}

func GetRecentSubmissionContents(assignment *model.Assignment, reference *model.ParsedCourseUserReference) (map[string]*model.GradingResult, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetRecentSubmissionContents(assignment, reference)
}

func RemoveSubmission(assignment *model.Assignment, email string, submissionID string) (bool, error) {
	if backend == nil {
		return false, fmt.Errorf("Database has not been opened.")
	}

	shortSubmissionID := common.GetShortSubmissionID(submissionID)
	return backend.RemoveSubmission(assignment, email, shortSubmissionID)
}

func GetSubmissionAttempts(assignment *model.Assignment, email string) ([]*model.GradingResult, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetSubmissionAttempts(assignment, email)
}
