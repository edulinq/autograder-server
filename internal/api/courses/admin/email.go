package admin

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/model"
)

type EmailRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	model.CourseMessageRecipients
	email.MessageContent

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

	recipients, err := request.CourseMessageRecipients.ToMessageRecipients()
	if err != nil {
		return nil, core.NewInternalError("-628", request, "Failed to resolve email recipients.").Err(err)
	}

	if recipients.IsEmpty() {
		return nil, core.NewBadRequestError("-629", request, "No email recipients provided.")
	}

	if !request.DryRun {
		err = email.SendFull(recipients.To, recipients.CC, recipients.BCC, request.Subject, request.Body, request.HTML)
		if err != nil {
			return nil, core.NewInternalError("-630", request, "Failed to send email.")
		}
	}

	response := EmailResponse{
		To:  recipients.To,
		CC:  recipients.CC,
		BCC: recipients.BCC,
	}

	return &response, nil
}
