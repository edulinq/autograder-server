package core

import (
	"context"

	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type GradeableRequest interface {
	GetContext() context.Context
	GetAssignment() *model.Assignment
	GetUserEmail() string
	GetMessage() string
	ToLogAttrs() []any
}
type BaseSubmitResponse struct {
	Rejected bool   `json:"rejected"`
	Message  string `json:"message"`

	GradingSuccess bool               `json:"grading-success"`
	GradingInfo    *model.GradingInfo `json:"result"`
}

func GradeToSubmissionResponse(request GradeableRequest, submissionPath string, options grader.GradeOptions) BaseSubmitResponse {
	response := BaseSubmitResponse{}

	result, reject, failureMessage, err := grader.Grade(
		request.GetContext(),
		request.GetAssignment(),
		submissionPath,
		request.GetUserEmail(),
		request.GetMessage(),
		options,
	)
	if err != nil {
		stdout := ""
		stderr := ""

		if (result != nil) && (result.HasTextOutput()) {
			stdout = result.Stdout
			stderr = result.Stderr
		}

		logAttributes := append([]any{err, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr)}, request.ToLogAttrs()...)
		log.Warn("Submission failed internally.", logAttributes)

		return response
	}

	if reject != nil {
		logAttributes := append([]any{log.NewAttr("reason", reject.String())}, request.ToLogAttrs()...)
		log.Debug("Submission rejected.", logAttributes)

		response.Rejected = true
		response.Message = reject.String()
		return response
	}

	if failureMessage != "" {
		logAttributes := append([]any{log.NewAttr("message", failureMessage)}, request.ToLogAttrs()...)
		log.Debug("Submission got a soft error.", logAttributes)

		response.Message = failureMessage
		return response
	}

	response.GradingSuccess = true
	response.GradingInfo = result.Info

	return response
}
