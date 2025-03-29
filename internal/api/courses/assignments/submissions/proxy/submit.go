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
	core.MinCourseRoleAdmin
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
	gradeOptions.ProxyUser = request.User.Email

	// If the proxy time is not specified, fall back to the assignment's due date.
	if request.ProxyTime == nil {
		// If the assignment does not have a due date, default to the Unix Epoch.
		if request.Assignment.DueDate == nil {
			proxyTime := timestamp.Zero()
			gradeOptions.ProxyTime = &proxyTime
		} else {
			gradeOptions.ProxyTime = request.Assignment.DueDate
		}
	} else {
		gradeOptions.ProxyTime = request.ProxyTime
	}

	// Proxy submissions are not subject to submission restrictions.
	result, reject, failureMessage, err := grader.Grade(request.Context, request.Assignment, request.Files.TempDir, request.ProxyUser.Email, request.Message, false, gradeOptions)
	if err != nil {
		stdout := ""
		stderr := ""

		if (result != nil) && (result.HasTextOutput()) {
			stdout = result.Stdout
			stderr = result.Stderr
		}

		log.LogToSplitLevels(log.LevelDebug, log.LevelInfo, "Submission failed internally.", err, request.Assignment, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr), request.User)

		return &response, nil
	}

	// A proxy submission should never be rejected.
	if reject != nil {
		log.LogToSplitLevels(log.LevelInfo, log.LevelError, "Proxy submission rejected.", request.Assignment, log.NewAttr("reason", reject.String()), log.NewAttr("request", request), request.User)

		response.Rejected = true
		response.Message = reject.String()
		return &response, nil
	}

	if failureMessage != "" {
		log.LogToSplitLevels(log.LevelTrace, log.LevelDebug, "Submission got a soft error.", request.Assignment, log.NewAttr("message", failureMessage), log.NewAttr("request", request), request.User)

		response.Message = failureMessage
		return &response, nil
	}

	response.GradingSuccess = true
	response.GradingInfo = result.Info

	return &response, nil
}
