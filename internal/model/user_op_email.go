package model

import (
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/util"
)

// Get the email message that corresponds with this user operation
// (or nil if there is no message to send).
func (this *UserOpResult) GetEmail() *email.Message {
	if this.Email == "" {
		return nil
	}

	if this.Added {
		return this.getAddEmail()
	}

	if this.Modified && (len(this.Enrolled) > 0) {
		return this.getEnrolledEmail()
	}

	if this.Modified && (this.CleartextPassword != "") {
		return this.getNewPasswordEmail()
	}

	return nil
}

func (this *UserOpResult) getAddEmail() *email.Message {
	body := fmt.Sprintf(baseAddBody, this.Email)

	if this.CleartextPassword != "" {
		body += fmt.Sprintf("\nYour new password token is '%s' (no quotes).\n", this.CleartextPassword)
	}

	return &email.Message{
		MessageRecipients: email.MessageRecipients{
			To: []string{this.Email},
		},
		MessageContent: email.MessageContent{
			Subject: fmt.Sprintf("Autograder %s -- User Account Created", config.NAME.Get()),
			Body:    body,
			HTML:    false,
		},
	}
}

func (this *UserOpResult) getEnrolledEmail() *email.Message {
	body := fmt.Sprintf(baseEnrollBody, this.Email, util.JoinStrings(", ", this.Enrolled...))

	return &email.Message{
		MessageRecipients: email.MessageRecipients{
			To: []string{this.Email},
		},
		MessageContent: email.MessageContent{
			Subject: fmt.Sprintf("Autograder %s -- Enrolled in Course", config.NAME.Get()),
			Body:    body,
			HTML:    false,
		},
	}
}

func (this *UserOpResult) getNewPasswordEmail() *email.Message {
	body := fmt.Sprintf(baseNewPassBody, this.Email, this.CleartextPassword)

	return &email.Message{
		MessageRecipients: email.MessageRecipients{
			To: []string{this.Email},
		},
		MessageContent: email.MessageContent{
			Subject: fmt.Sprintf("Autograder %s -- New Password Token", config.NAME.Get()),
			Body:    body,
			HTML:    false,
		},
	}
}

var baseAddBody string = `Hello,

An autograder account with the username/email '%s' has been created.
Usage instructions will be provided in class.
`

var baseEnrollBody string = `Hello,

Your autograder account '%s' has been enrolled in the course(s) '%s'.
Usage instructions will be provided in class.
`

var baseNewPassBody string = `Hello,

Your autograder account '%s' has been generated a new password token: '%s' (no quotes).
`
