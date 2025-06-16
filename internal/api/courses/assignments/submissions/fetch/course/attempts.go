package course

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type FetchCourseAttemptsRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	TargetUsers []model.CourseUserReference `json:"target-users"`
}

type FetchCourseAttemptsResponse struct {
	GradingResults map[string]*model.GradingResult `json:"grading-results"`
}

// Get all recent submissions and grading information for this assignment.
func HandleFetchCourseAttempts(request *FetchCourseAttemptsRequest) (*FetchCourseAttemptsResponse, *core.APIError) {
	// Default to getting the most recent submissions and grading information for all users in the course.
	if len(request.TargetUsers) == 0 {
		request.TargetUsers = []model.CourseUserReference{"*"}
	}

	reference, err := model.ParseCourseUserReferences(request.TargetUsers)
	if err != nil {
		return nil, core.NewBadRequestError("-637", request, "Failed to parse target users.").Err(err)
	}

	results, err := db.GetRecentSubmissionContents(request.Assignment, *reference)
	if err != nil {
		return nil, core.NewInternalError("-605", request, "Failed to get submissions.").Err(err)
	}

	return &FetchCourseAttemptsResponse{results}, nil
}
