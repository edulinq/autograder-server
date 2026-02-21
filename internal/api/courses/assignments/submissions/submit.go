package submissions

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/grader"
)

type SubmitRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleStudent
	Files core.POSTFiles `json:"-"`

	Message   string `json:"message"`
	AllowLate bool   `json:"allow-late"`
}

type SubmitResponse struct {
	core.BaseSubmitResponse
}

// Submit an assignment submission to the autograder.
func HandleSubmit(request *SubmitRequest) (*SubmitResponse, *core.APIError) {
	gradeOptions := grader.GetDefaultGradeOptions()
	gradeOptions.AllowLate = request.AllowLate

	base, apiErr := core.GradeRequestSubmission(request.APIRequestAssignmentContext, request.Files.TempDir, request.User.Email, request.Message, gradeOptions)
	if apiErr != nil {
		return nil, apiErr
	}

	return &SubmitResponse{BaseSubmitResponse: *base}, nil
}
