package admin

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
)

type SendEmailRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	To      []string `json:"to"`

	// Do not actually send any emails.
	DryRun bool `json:"dry-run"`
}

type SendEmailResponse struct {
	Success    bool     `json:"success"`
	Recipients []string `json:"recipients"`
}

// Send an email to course user(s).
// TODO:
// tests
// bcc and cc (no infrastructure)
func HandleSendEmail(request *SendEmailRequest) (*SendEmailResponse, *core.APIError) {
	response := SendEmailResponse{}

	var err error
	request.To, err = db.ResolveCourseUsers(request.Course, request.To)
	if err != nil {
		return nil, core.NewInternalError("-627", &request.APIRequestCourseUserContext, "Failed to resolve users.")
	}

	if !request.DryRun {
		err = email.Send(request.To, request.Subject, request.Body, false)
		if err != nil {
			log.Error("Failed to send email(s).", err, request)
		}
	}

	response.Success = (err == nil)
	response.Recipients = request.To

	return &response, nil
}
