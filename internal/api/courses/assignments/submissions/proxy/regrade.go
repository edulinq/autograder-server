package proxy

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

type RegradeRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	grader.RegradeOptions
}

type RegradeResponse struct {
	Complete      bool                                    `json:"complete"`
	Options       grader.RegradeOptions                   `json:"options"`
	ResolvedUsers []string                                `json:"resolved-users"`
	RegradeAfter  timestamp.Timestamp                     `json:"regrade-after"`
	Results       map[string]*model.SubmissionHistoryItem `json:"results"`
	WorkErrors    map[string]string                       `json:"work-errors"`
}

func HandleRegrade(request *RegradeRequest) (*RegradeResponse, *core.APIError) {
	gradeOptions := grader.GetDefaultGradeOptions()
	// Proxy submissions are not subject to submission restrictions.
	gradeOptions.CheckRejection = false
	gradeOptions.ProxyUser = request.User.Email

	request.RegradeOptions.GradeOptions = gradeOptions

	results, regradeAfter, numRemaining, workErrors, err := grader.Regrade(request.Assignment, request.RegradeOptions)
	if err != nil {
		return nil, core.NewInternalError("-638", request, "Failed to get courses from database.").Err(err)
	}

	response := RegradeResponse{
		// TODO: Resolved users may have users from outside the course (so it will never be complete).
		Complete:      (numRemaining == len(request.RegradeOptions.ResolvedUsers)),
		Options:       request.RegradeOptions,
		ResolvedUsers: request.RegradeOptions.ResolvedUsers,
		RegradeAfter:  regradeAfter,
		Results:       results,
		WorkErrors:    workErrors,
	}

	return &response, nil
}
