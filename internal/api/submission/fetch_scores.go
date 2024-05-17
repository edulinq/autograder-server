package submission

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchScoresRequest struct {
	core.APIRequestAssignmentContext
	core.MinRoleGrader

	// Filter results to only users with this role.
	FilterRole model.UserRole `json:"filter-role"`
}

type FetchScoresResponse struct {
	SubmissionInfos map[string]*model.SubmissionHistoryItem `json:"submission-infos"`
}

func HandleFetchScores(request *FetchScoresRequest) (*FetchScoresResponse, *core.APIError) {
	submissionInfos, err := db.GetRecentSubmissionSurvey(request.Assignment, request.FilterRole)
	if err != nil {
		return nil, core.NewInternalError("-602", &request.APIRequestCourseUserContext, "Failed to get submission summaries.").
			Err(err).Assignment(request.Assignment.GetID())
	}

	return &FetchScoresResponse{submissionInfos}, nil
}
