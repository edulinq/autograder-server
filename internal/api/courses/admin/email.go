package admin

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type EmailRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader
	Users core.CourseUsers `json:"-"`

	model.CourseMessageRecipients
	email.MessageContent

	DryRun bool `json:"dry-run"`
}

type EmailResponse struct {
	To  []string `json:"to"`
	CC  []string `json:"cc"`
	BCC []string `json:"bcc"`

	Errors map[string]string `json:"errors,omitempty,omitzero"`
}

// Send an email to course users.
func HandleEmail(request *EmailRequest) (*EmailResponse, *core.APIError) {
	if request.Subject == "" {
		return nil, core.NewBadRequestError("-627", request, "No email subject provided.")
	}

	recipients, userErrors := request.CourseMessageRecipients.ToMessageRecipients(request.Users)

	errors := make(map[string]string, len(userErrors))

	for reference, err := range userErrors {
		errors[reference] = err.Error()

		log.Warn("Failed to parse user reference.", err, log.NewAttr("reference", reference))
	}

	if len(errors) != 0 {
		return &EmailResponse{
			To:     nil,
			CC:     nil,
			BCC:    nil,
			Errors: errors,
		}, nil
	}

	if recipients.IsEmpty() {
		return nil, core.NewBadRequestError("-628", request, "No email recipients provided.")
	}

	if !request.DryRun {
		err := email.SendFull(recipients.To, recipients.CC, recipients.BCC, request.Subject, request.Body, request.HTML)
		if err != nil {
			return nil, core.NewInternalError("-629", request, "Failed to send email.")
		}
	}

	response := EmailResponse{
		To:     recipients.To,
		CC:     recipients.CC,
		BCC:    recipients.BCC,
		Errors: nil,
	}

	return &response, nil
}
