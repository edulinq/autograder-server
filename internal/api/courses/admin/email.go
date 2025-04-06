package admin

import (
	"errors"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
)

type EmailRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	email.Message

	DryRun bool `json:"dry-run"`
}

type EmailResponse struct {
	To  []string `json:"to"`
	CC  []string `json:"cc"`
	BCC []string `json:"bcc"`
}

// Send an email to course users.
func HandleEmail(request *EmailRequest) (*EmailResponse, *core.APIError) {
	if request.Subject == "" {
		return nil, core.NewBadRequestError("-627", request, "No email subject provided.")
	}

	var err error
	var errs error

	request.To, err = db.ResolveCourseUsers(request.Course, request.To)
	errs = errors.Join(errs, err)

	request.CC, err = db.ResolveCourseUsers(request.Course, request.CC)
	errs = errors.Join(errs, err)

	request.BCC, err = db.ResolveCourseUsers(request.Course, request.BCC)
	errs = errors.Join(errs, err)

	if errs != nil {
		return nil, core.NewInternalError("-628", request, "Failed to resolve email recipients.").Err(errs)
	}

	if (len(request.To) + len(request.CC) + len(request.BCC)) == 0 {
		return nil, core.NewBadRequestError("-629", request, "No email recipients provided.")
	}

	if !request.DryRun {
		err = email.SendFull(request.To, request.CC, request.BCC, request.Subject, request.Body, request.HTML)
		if err != nil {
			return nil, core.NewInternalError("-630", request, "Failed to send email.")
		}
	}

	response := EmailResponse{
		To:  request.To,
		CC:  request.CC,
		BCC: request.BCC,
	}

	return &response, nil
}
