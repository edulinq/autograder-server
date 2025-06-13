package model

import (
	"errors"

	"github.com/edulinq/autograder/internal/email"
)

type ServerMessageRecipients struct {
	To  []ServerUserReference `json:"to"`
	CC  []ServerUserReference `json:"cc"`
	BCC []ServerUserReference `json:"bcc"`
}

type CourseMessageRecipients struct {
	To  []CourseUserReference `json:"to"`
	CC  []CourseUserReference `json:"cc"`
	BCC []CourseUserReference `json:"bcc"`
}

func (this *ServerMessageRecipients) ToMessageRecipients(courses map[string]*Course, users map[string]*ServerUser) (*email.MessageRecipients, error) {
	var errs error = nil

	reference, err := ParseServerUserReferences(this.To, courses)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	to := ResolveServerUserEmails(users, reference)

	reference, err = ParseServerUserReferences(this.CC, courses)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	cc := ResolveServerUserEmails(users, reference)

	reference, err = ParseServerUserReferences(this.BCC, courses)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	bcc := ResolveServerUserEmails(users, reference)

	recipients := email.MessageRecipients{
		To:  to,
		CC:  cc,
		BCC: bcc,
	}

	return &recipients, errs
}

func (this *CourseMessageRecipients) ToMessageRecipients(users map[string]*CourseUser) (*email.MessageRecipients, error) {
	var errs error = nil

	reference, err := ParseCourseUserReferences(this.To)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	to := ResolveCourseUserEmails(users, reference)

	reference, err = ParseCourseUserReferences(this.CC)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	cc := ResolveCourseUserEmails(users, reference)

	reference, err = ParseCourseUserReferences(this.BCC)
	if err != nil {
		errs = errors.Join(errs, err)
	}

	bcc := ResolveCourseUserEmails(users, reference)

	recipients := email.MessageRecipients{
		To:  to,
		CC:  cc,
		BCC: bcc,
	}

	return &recipients, errs
}
