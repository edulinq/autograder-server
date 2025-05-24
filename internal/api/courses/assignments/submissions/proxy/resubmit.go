package proxy

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type ResubmitRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader

	TargetSubmission string `json:"target-submission"`

	ProxyUser core.TargetCourseUser `json:"proxy-email"`
	ProxyTime *timestamp.Timestamp  `json:"proxy-time"`
}

type ResubmitResponse struct {
	core.BaseSubmitResponse

	FoundUser       bool `json:"found-user"`
	FoundSubmission bool `json:"found-submission"`
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
		return nil, core.NewInternalError("-631", request, "Failed to get submission contents.").
			Err(err).Add("target-user", request.ProxyUser.Email).Add("submission", request.TargetSubmission)
	}

	if gradingResult == nil {
		return &response, nil
	}

	response.FoundSubmission = true

	tempDir, err := util.MkDirTemp("resumbit-request-files-")
	if err != nil {
		return nil, core.NewInternalError("-632", request, "Failed to create temp resubmit files directory.").
			Err(err).Add("target-user", request.ProxyUser.Email).Add("submission", request.TargetSubmission)
	}
	defer util.RemoveDirent(tempDir)

	err = util.GzipBytesToDirectory(tempDir, gradingResult.InputFilesGZip)
	if err != nil {
		return nil, core.NewInternalError("-633", request, "Failed to write submission input to a temp dir.").
			Err(err).Add("target-user", request.ProxyUser.Email).Add("submission", request.TargetSubmission)
	}

	message := ""
	if gradingResult.Info != nil {
		message = gradingResult.Info.Message
	}

	gradeOptions := grader.GetDefaultGradeOptions()
	// Proxy submissions are not subject to submission restrictions.
	gradeOptions.CheckRejection = false
	gradeOptions.ProxyUser = request.User.Email
	gradeOptions.ProxyTime = grader.ResolveProxyTime(request.ProxyTime, request.Assignment)

	response.BaseSubmitResponse = core.GradeRequestSubmission(request.APIRequestAssignmentContext, tempDir, request.ProxyUser.Email, message, gradeOptions)

	return &response, nil
}
