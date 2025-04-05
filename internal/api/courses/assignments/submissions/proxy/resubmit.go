package proxy

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/log"
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

type ResubmitResponse struct {
	FoundUser       bool   `json:"found-user"`
	FoundSubmission bool   `json:"found-submission"`
	Rejected        bool   `json:"rejected"`
	Message         string `json:"message"`

	GradingSuccess bool               `json:"grading-success"`
	GradingInfo    *model.GradingInfo `json:"result"`
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

	err = util.GzipBytesToDirectory(tempDir, gradingResult.InputFilesGZip)
	if err != nil {
		util.RemoveDirent(tempDir)
		return nil, core.NewInternalError("-634", &request.APIRequestCourseUserContext, "Failed to write submission input to a temp dir.").
			Err(err).Assignment(request.Assignment.GetID()).
			Add("target-user", request.ProxyUser.Email).Add("submission", request.TargetSubmission)
	}

	gradeOptions := grader.GetDefaultGradeOptions()
	// Proxy submissions are not subject to submission restrictions.
	gradeOptions.CheckRejection = false
	gradeOptions.ProxyUser = request.User.Email
	gradeOptions.ProxyTime = grader.ResolveProxyTime(request.ProxyTime, request.Assignment)

	result, reject, failureMessage, err := grader.Grade(request.Context, request.Assignment, tempDir, request.ProxyUser.Email, request.Message, gradeOptions)
	if err != nil {
		stdout := ""
		stderr := ""

		if (result != nil) && (result.HasTextOutput()) {
			stdout = result.Stdout
			stderr = result.Stderr
		}

		log.Warn("Resubmission failed internally.", err, request.Assignment, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr), request.User)

		return &response, nil
	}

	// A proxy resubmission should never be rejected.
	if reject != nil {
		log.Error("Proxy resubmission rejected.", request.Assignment, log.NewAttr("reason", reject.String()), log.NewAttr("request", request), request.User)

		response.Rejected = true
		response.Message = reject.String()
		return &response, nil
	}

	if failureMessage != "" {
		log.Debug("Resubmission got a soft error.", request.Assignment, log.NewAttr("message", failureMessage), log.NewAttr("request", request), request.User)

		response.Message = failureMessage
		return &response, nil
	}

	response.GradingSuccess = true
	response.GradingInfo = result.Info

	return &response, nil
}
