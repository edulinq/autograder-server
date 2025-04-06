package proxy

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/grader"
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

type SubmitResponse struct {
	core.BaseSubmitResponse

	FoundUser bool `json:"found-user"`
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

	response.BaseSubmitResponse = core.GradeToSubmissionResponse(request.APIRequestAssignmentContext, request.Files.TempDir, request.ProxyUser.Email, request.Message, gradeOptions)

	return &response, nil
}
