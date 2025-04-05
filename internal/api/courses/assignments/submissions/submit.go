package submissions

import (
	"context"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
)

type SubmitRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleStudent
	Files core.POSTFiles `json:"-"`

	Message   string `json:"message"`
	AllowLate bool   `json:"allow-late"`
}

func (request SubmitRequest) GetContext() context.Context {
	return request.Context
}

func (request SubmitRequest) GetAssignment() *model.Assignment {
	return request.Assignment
}

func (request SubmitRequest) GetUserEmail() string {
	return request.User.Email
}

func (request SubmitRequest) GetMessage() string {
	return request.Message
}

func (request SubmitRequest) ToLogAttrs() []any {
	return core.GetLogAttributesFromAPIRequest(&request)
}

type SubmitResponse struct {
	core.BaseSubmitResponse
}

// Submit an assignment submission to the autograder.
func HandleSubmit(request *SubmitRequest) (*SubmitResponse, *core.APIError) {
	response := SubmitResponse{}

	gradeOptions := grader.GetDefaultGradeOptions()
	gradeOptions.AllowLate = request.AllowLate

	response.BaseSubmitResponse = core.GradeToSubmissionResponse(request, request.Files.TempDir, gradeOptions)

	return &response, nil
}
