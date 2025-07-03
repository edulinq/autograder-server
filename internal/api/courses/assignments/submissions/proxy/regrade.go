package proxy

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
)

type RegradeRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	grader.RegradeOptions
}

type RegradeResponse struct {
	grader.RegradeResult

	Complete      bool     `json:"complete"`
	ResolvedUsers []string `json:"resolved-users"`
}

// Proxy regrade an assignment for all target users using their most recent submission.
func HandleRegrade(request *RegradeRequest) (*RegradeResponse, *core.APIError) {
	if len(request.RegradeOptions.RawReferences) == 0 {
		request.RegradeOptions.RawReferences = model.NewAllCourseUserReference()
	}

	gradeOptions := grader.GetDefaultGradeOptions()
	// Proxy submissions are not subject to submission restrictions.
	gradeOptions.CheckRejection = false
	gradeOptions.ProxyUser = request.User.Email

	request.RegradeOptions.GradeOptions = gradeOptions
	request.RegradeOptions.Context = request.APIRequestUserContext.Context

	result, numRemaining, err := grader.Regrade(request.Assignment, request.RegradeOptions)
	if err != nil {
		return nil, core.NewInternalError("-638", request, "Failed to get courses from database.").Err(err)
	}

	response := RegradeResponse{
		RegradeResult: *result,
		Complete:      (numRemaining == 0),
		ResolvedUsers: result.Options.ResolvedUsers,
	}

	return &response, nil
}
