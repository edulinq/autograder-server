package course

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchCourseScoresRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	// Filter results to only users with this role.
	FilterRole model.CourseUserRole `json:"filter-role"`
}

type FetchCourseScoresResponse struct {
	SubmissionInfos map[string]*model.SubmissionHistoryItem `json:"submission-infos"`
}

// Get a summary of the most recent scores for this assignment.
func HandleFetchCourseScores(request *FetchCourseScoresRequest) (*FetchCourseScoresResponse, *core.APIError) {
	submissionInfos, err := db.GetRecentSubmissionSurvey(request.Assignment, request.FilterRole)
	if err != nil {
		return nil, core.NewInternalError("-602", request, "Failed to get submission summaries.").Err(err)
	}

	return &FetchCourseScoresResponse{submissionInfos}, nil
}
