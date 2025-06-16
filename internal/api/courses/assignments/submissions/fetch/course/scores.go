package course

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchCourseScoresRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	// Filter results to users that match the course user references.
	// If TargetUsers is empty, results for all users in the course will be fetched.
	TargetUsers []model.CourseUserReference `json:"target-users"`
}

type FetchCourseScoresResponse struct {
	SubmissionInfos map[string]*model.SubmissionHistoryItem `json:"submission-infos"`
}

// Get a summary of the most recent scores for this assignment.
func HandleFetchCourseScores(request *FetchCourseScoresRequest) (*FetchCourseScoresResponse, *core.APIError) {
	// Default to getting the most recent scores for all users in the course.
	if len(request.TargetUsers) == 0 {
		request.TargetUsers = []model.CourseUserReference{"*"}
	}

	reference, err := model.ParseCourseUserReferences(request.TargetUsers)
	if err != nil {
		return nil, core.NewBadRequestError("-636", request, "Failed to parse target users.").Err(err)
	}

	submissionInfos, err := db.GetRecentSubmissionSurvey(request.Assignment, *reference)
	if err != nil {
		return nil, core.NewInternalError("-602", request, "Failed to get submission summaries.").Err(err)
	}

	return &FetchCourseScoresResponse{submissionInfos}, nil
}
