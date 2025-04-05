package proxy

import (
	"context"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type ResubmitRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader
	TargetSubmission string `json:"target-submission"`

	ProxyUser core.TargetCourseUser `json:"proxy-email"`
	ProxyTime *timestamp.Timestamp  `json:"proxy-time"`

	Message string `json:"message"`
}

func (request ResubmitRequest) GetContext() context.Context {
	return request.Context
}

func (request ResubmitRequest) GetAssignment() *model.Assignment {
	return request.Assignment
}

func (request ResubmitRequest) GetUserEmail() string {
	return request.ProxyUser.Email
}

func (request ResubmitRequest) GetMessage() string {
	return request.Message
}

func (request ResubmitRequest) ToLogAttrs() []any {
	return core.GetLogAttributesFromAPIRequest(&request)
}

type ResubmitResponse struct {
	FoundUser       bool `json:"found-user"`
	FoundSubmission bool `json:"found-submission"`

	core.BaseSubmitResponse
}

// Proxy resubmit an assignment submission to the autograder.
func HandleResubmit(request *ResubmitRequest) (*ResubmitResponse, *core.APIError) {
	response := ResubmitResponse{}

	if !request.ProxyUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	gradingResult, err := db.GetSubmissionContents(request.Assignment, request.ProxyUser.Email, request.TargetSubmission)
	if err != nil {
		return nil, core.NewInternalError("-632", &request.APIRequestCourseUserContext, "Failed to get submission contents.").
			Err(err).Assignment(request.Assignment.GetID()).
			Add("target-user", request.ProxyUser.Email).Add("submission", request.TargetSubmission)
	}

	if gradingResult == nil {
		return &response, nil
	}

	response.FoundSubmission = true

	tempDir, err := util.MkDirTemp("resumbit-request-files-")
	if err != nil {
		return nil, core.NewInternalError("-633", &request.APIRequestCourseUserContext, "Failed to create temp resubmit files directory.").
			Err(err).Assignment(request.Assignment.GetID()).
			Add("target-user", request.ProxyUser.Email).Add("submission", request.TargetSubmission)
	}
	defer util.RemoveDirent(tempDir)

	err = util.GzipBytesToDirectory(tempDir, gradingResult.InputFilesGZip)
	if err != nil {
		return nil, core.NewInternalError("-634", &request.APIRequestCourseUserContext, "Failed to write submission input to a temp dir.").
			Err(err).Assignment(request.Assignment.GetID()).
			Add("target-user", request.ProxyUser.Email).Add("submission", request.TargetSubmission)
	}

	gradeOptions := grader.GetDefaultGradeOptions()
	// Proxy submissions are not subject to submission restrictions.
	gradeOptions.CheckRejection = false
	gradeOptions.ProxyUser = request.User.Email
	gradeOptions.ProxyTime = grader.ResolveProxyTime(request.ProxyTime, request.Assignment)

	response.BaseSubmitResponse = core.GradeToSubmissionResponse(request, tempDir, gradeOptions)

	return &response, nil
}
