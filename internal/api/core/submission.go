package core

import (
	"fmt"

	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type BaseSubmitResponse struct {
	Rejected bool   `json:"rejected"`
	Message  string `json:"message"`

	GradingSuccess bool               `json:"grading-success"`
	GradingInfo    *model.GradingInfo `json:"result"`
}

func GradeRequestSubmission(request APIRequestAssignmentContext, submissionPath string, email string, message string, options grader.GradeOptions) BaseSubmitResponse {
	response := BaseSubmitResponse{}

	result, reject, failureMessage, err := grader.Grade(
		request.Context,
		request.Assignment,
		submissionPath,
		email,
		message,
		options,
	)
	if err != nil {
		stdout := ""
		stderr := ""

		if (result != nil) && (result.HasTextOutput()) {
			stdout = result.Stdout
			stderr = result.Stderr
		}

		logAttributes := append([]any{err, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr)}, getLogAttributesFromAPIRequest(&request)...)
		log.Warn("Submission failed internally.", logAttributes...)

		return response
	}

	if reject != nil {
		logAttributes := append([]any{log.NewAttr("reason", reject.String())}, getLogAttributesFromAPIRequest(&request)...)
		log.Debug("Submission rejected.", logAttributes...)

		response.Rejected = true
		response.Message = reject.String()
		return response
	}

	if failureMessage != "" {
		stdout := ""
		stderr := ""

		if (result != nil) && (result.HasTextOutput()) {
			stdout = result.Stdout
			stderr = result.Stderr
		}

		logAttributes := append([]any{log.NewAttr("message", failureMessage), log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr)}, getLogAttributesFromAPIRequest(&request)...)
		log.Debug("Submission got a soft error.", logAttributes...)

		response.Message = ConcatStdOutErr(failureMessage, stdout, stderr)
		return response
	}

	response.GradingSuccess = true
	response.GradingInfo = result.Info

	return response
}

// Concat stdout/stderr to the given message.
func ConcatStdOutErr(message string, stdout string, stderr string) string {
	message += fmt.Sprintf("\n--- stdout ---\n%s\n--------------\n", stdout)
	message += fmt.Sprintf("\n--- stderr ---\n%s\n--------------\n", stderr)
	return message
}
