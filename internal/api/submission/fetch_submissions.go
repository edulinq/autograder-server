package submission

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchSubmissionsRequest struct {
	core.APIRequestAssignmentContext
	core.MinRoleGrader

	FilterRole model.UserRole `json:"filter-role"`
}

type FetchSubmissionsResponse struct {
	GradingResults map[string]*model.GradingResult `json:"grading-results"`
}

func HandleFetchSubmissions(request *FetchSubmissionsRequest) (*FetchSubmissionsResponse, *core.APIError) {
	results, err := db.GetRecentSubmissionContents(request.Assignment, request.FilterRole)
	if err != nil {
		return nil, core.NewInternalError("-605", &request.APIRequestCourseUserContext, "Failed to get submissions.").
			Err(err).Assignment(request.Assignment.GetID())
	}

	return &FetchSubmissionsResponse{results}, nil
}
