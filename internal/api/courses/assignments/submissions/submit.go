package submissions

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type SubmitRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleStudent
	Files core.POSTFiles

	Message string `json:"message"`
	AckLate bool   `json:"ack-late"`
}

type SubmitResponse struct {
	Rejected bool   `json:"rejected"`
	Message  string `json:"message"`

	GradingSucess bool               `json:"grading-success"`
	GradingInfo   *model.GradingInfo `json:"result"`
}

// Submit an assignment submission to the autograder.
func HandleSubmit(request *SubmitRequest) (*SubmitResponse, *core.APIError) {
	response := SubmitResponse{}

	result, reject, failureMessage, err := grader.GradeDefault(request.Assignment, request.Files.TempDir, request.User.Email, request.Message, request.AckLate)
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

	if reject != nil {
		log.LogToSplitLevels(log.LevelTrace, log.LevelDebug, "Submission rejected.", request.Assignment, log.NewAttr("reason", reject.String()), log.NewAttr("request", request), request.User)

		response.Rejected = true
		response.Message = reject.String()
		return &response, nil
	}

	if failureMessage != "" {
		log.LogToSplitLevels(log.LevelTrace, log.LevelDebug, "Submission got a soft error.", request.Assignment, log.NewAttr("message", failureMessage), log.NewAttr("request", request), request.User)

		response.Message = failureMessage
		return &response, nil
	}

	response.GradingSucess = true
	response.GradingInfo = result.Info

	return &response, nil
}
