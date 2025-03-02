package admin

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
)

type EmailRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader

	email.Message

	DryRun bool `json:"dry-run"`
}

type EmailResponse struct {
	Success bool     `json:"success"`
	To      []string `json:"to"`
	CC      []string `json:"cc"`
	BCC     []string `json:"bcc"`
}

// Send an email to course user(s).
func HandleEmail(request *EmailRequest) (*EmailResponse, *core.APIError) {
	response := EmailResponse{}
	var err error

	if request.Subject == "" {
		return nil, core.NewBadRequestError("-627", &request.APIRequest, "No subject.")
	}

	for _, section := range []string{"to", "cc", "bcc"} {
		switch section {
		case "to":
			request.To, err = db.ResolveCourseUsers(request.Course, request.To)
		case "cc":
			request.CC, err = db.ResolveCourseUsers(request.Course, request.CC)
		case "bcc":
			request.BCC, err = db.ResolveCourseUsers(request.Course, request.BCC)
		}

		if err != nil {
			return nil, core.NewInternalError("-628", &request.APIRequestCourseUserContext, "Failed to resolve '"+section+"' emails.")
		}
	}

	if (len(request.To) + len(request.CC) + len(request.BCC)) == 0 {
		return nil, core.NewBadRequestError("-629", &request.APIRequest, "No recipients.")
	}

	if !request.DryRun {
		err = email.Send(request.To, request.Subject, request.Body, false)
		if err != nil {
			log.Error("Failed to send email(s).", err, request)
		}
	}

	response.Success = (err == nil)
	response.To = request.To
	response.CC = request.CC
	response.BCC = request.BCC

	return &response, nil
}
