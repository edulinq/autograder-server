package model

import (
	"github.com/edulinq/autograder/internal/email"
)

type CourseMessageRecipients struct {
	To  []CourseUserReference `json:"to"`
	CC  []CourseUserReference `json:"cc"`
	BCC []CourseUserReference `json:"bcc"`
}

func (this *CourseMessageRecipients) ToMessageRecipients(users map[string]*CourseUser) (*email.MessageRecipients, map[string]error) {
	userErrors := make(map[string]error, 0)

	reference, errs := ParseCourseUserReferences(this.To)
	for input, err := range errs {
		userErrors[input] = err
	}

	to := ResolveCourseUserEmails(users, reference)

	reference, errs = ParseCourseUserReferences(this.CC)
	for input, err := range errs {
		userErrors[input] = err
	}

	cc := ResolveCourseUserEmails(users, reference)

	reference, errs = ParseCourseUserReferences(this.BCC)
	for input, err := range errs {
		userErrors[input] = err
	}

	bcc := ResolveCourseUserEmails(users, reference)

	recipients := email.MessageRecipients{
		To:  to,
		CC:  cc,
		BCC: bcc,
	}

	if len(userErrors) == 0 {
		userErrors = nil
	}

	return &recipients, userErrors
}
