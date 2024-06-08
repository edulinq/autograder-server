package model

import (
	"github.com/edulinq/autograder/internal/email"
)

// Get the email message that corresponds with this user operation
// (or nil if there is no message to send).
func (this *UserOpResult) GetEmail() *email.Message {
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
	// TEST
	return nil
}

func (this *UserOpResult) getEnrolledEmail() *email.Message {
	// TEST
	return nil
}

func (this *UserOpResult) getNewPasswordEmail() *email.Message {
	// TEST
	return nil
}
