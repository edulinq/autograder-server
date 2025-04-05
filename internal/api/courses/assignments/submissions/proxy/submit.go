package proxy

import (
	"context"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

type SubmitRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader
	Files core.POSTFiles `json:"-"`

	ProxyUser core.TargetCourseUser `json:"proxy-email"`
	ProxyTime *timestamp.Timestamp  `json:"proxy-time"`

	Message string `json:"message"`
}

func (request SubmitRequest) GetContext() context.Context {
	return request.Context
}

func (request SubmitRequest) GetAssignment() *model.Assignment {
	return request.Assignment
}

func (request SubmitRequest) GetUserEmail() string {
	return request.ProxyUser.Email
}

func (request SubmitRequest) GetMessage() string {
	return request.Message
}

func (request SubmitRequest) ToLogAttrs() []any {
	return core.GetLogAttributesFromAPIRequest(&request)
}

type SubmitResponse struct {
	FoundUser bool `json:"found-user"`

	core.BaseSubmitResponse
}

// Proxy submit an assignment submission to the autograder.
func HandleSubmit(request *SubmitRequest) (*SubmitResponse, *core.APIError) {
	response := SubmitResponse{}

	if !request.ProxyUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	gradeOptions := grader.GetDefaultGradeOptions()
	// Proxy submissions are not subject to submission restrictions.
	gradeOptions.CheckRejection = false
	gradeOptions.ProxyUser = request.User.Email
	gradeOptions.ProxyTime = grader.ResolveProxyTime(request.ProxyTime, request.Assignment)

	response.BaseSubmitResponse = core.GradeToSubmissionResponse(request, request.Files.TempDir, gradeOptions)

	return &response, nil
}
