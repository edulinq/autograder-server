package submissions

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchCourseAttemptsRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	FilterRole model.CourseUserRole `json:"filter-role"`
}

type FetchCourseAttemptsResponse struct {
	GradingResults map[string]*model.GradingResult `json:"grading-results"`
}

func HandleFetchCourseAttempts(request *FetchCourseAttemptsRequest) (*FetchCourseAttemptsResponse, *core.APIError) {
	results, err := db.GetRecentSubmissionContents(request.Assignment, request.FilterRole)
	if err != nil {
		return nil, core.NewInternalError("-605", &request.APIRequestCourseUserContext, "Failed to get submissions.").
			Err(err).Assignment(request.Assignment.GetID())
	}

	return &FetchCourseAttemptsResponse{results}, nil
}
