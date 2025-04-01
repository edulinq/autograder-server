package proxy

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/log"
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

type SubmitResponse struct {
	FoundUser bool   `json:"found-user"`
	Rejected  bool   `json:"rejected"`
	Message   string `json:"message"`

	GradingSuccess bool               `json:"grading-success"`
	GradingInfo    *model.GradingInfo `json:"result"`
}

// Proxy submit an assignment submission to the autograder.
func HandleSubmit(request *SubmitRequest) (*SubmitResponse, *core.APIError) {
	response := SubmitResponse{}

	if !request.ProxyUser.Found {
		return &response, nil
	}

	gradeOptions := grader.GetDefaultGradeOptions()
	// Proxy submissions are not subject to submission restrictions.
	gradeOptions.CheckRejection = false
	gradeOptions.ProxyUser = request.User.Email
	gradeOptions.ProxyTime = grader.ResolveProxyTime(request.ProxyTime, request.Assignment)

	result, reject, failureMessage, err := grader.Grade(request.Context, request.Assignment, request.Files.TempDir, request.ProxyUser.Email, request.Message, gradeOptions)
	if err != nil {
		stdout := ""
		stderr := ""

		if (result != nil) && (result.HasTextOutput()) {
			stdout = result.Stdout
			stderr = result.Stderr
		}

		log.Warn("Submission failed internally.", err, request.Assignment, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr), request.User)

		return &response, nil
	}

	// A proxy submission should never be rejected.
	if reject != nil {
		log.Error("Proxy submission rejected.", request.Assignment, log.NewAttr("reason", reject.String()), log.NewAttr("request", request), request.User)

		response.Rejected = true
		response.Message = reject.String()
		return &response, nil
	}

	if failureMessage != "" {
		log.Debug("Submission got a soft error.", request.Assignment, log.NewAttr("message", failureMessage), log.NewAttr("request", request), request.User)

		response.Message = failureMessage
		return &response, nil
	}

	response.GradingSuccess = true
	response.GradingInfo = result.Info

	return &response, nil
}
