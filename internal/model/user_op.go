package model

import (
	"github.com/edulinq/autograder/internal/util"
)

// A general representation of the result of operations that modify a user in any way (add, remove, enroll, drop, etc).
// All user-facing functions (essentially non-db functions) should return an instance or collection of these objects.
type UserOpResult struct {
	// The email/id of the target user.
	Email string `json:"email"`

	// The user was added to the server.
	Added bool `json:"added,omitempty"`

	// The user existed before this operation and was edited (including enrollment changes).
	Modified bool `json:"modified,omitempty"`

	// The user existed before this operation and was removed.
	Removed bool `json:"removed,omitempty"`

	// The user was skipped (often because they already exist).
	Skipped bool `json:"skipped,omitempty"`

	// The user did not exist before this operation and does not exist after.
	// This may also be an error depending on the semantics of the operation.
	NotExists bool `json:"not-exists,omitempty"`

	// The user was emailed during the course of this operation.
	// This is more than just GetEmail() was called, an actual email was sent
	// (or would have been sent if this operation was during a dry-run).
	Emailed bool `json:"emailed,omitempty"`

	// The user was enrolled in the following courses (by id).
	Enrolled []string `json:"enrolled,omitempty"`

	// The user was removed from the following courses (by id).
	Dropped []string `json:"dropped,omitempty"`

	// The following error occurred during this operation because of the provided data,
	// i.e., they are caused by the calling user.
	// All error messages should be safe for users.
	ValidationErrors []string `json:"validation-errors,omitempty"`

	// The following error occurred during this operation, but not because of the provided,
	// i.e., they are the system's fault.
	// These errors are not guarenteed to be safe for users,
	// and the calling code should decide how they should be managed.
	SystemErrors []string `json:"system-errors,omitempty"`

	// The following cleartext password was generated during this operation.
	// Care should be taken to not expose their field.
	CleartextPassword string `json:"cleartext-password,omitempty"`
}

func NewValidationErrorUserOpResult(email string, err error) *UserOpResult {
	return &UserOpResult{
		Email:            email,
		ValidationErrors: []string{err.Error()},
	}
}

func NewSystemErrorUserOpResult(email string, err error) *UserOpResult {
	return &UserOpResult{
		Email:        email,
		SystemErrors: []string{err.Error()},
	}
}

func (this *UserOpResult) HasErrors() bool {
	return (len(this.ValidationErrors) > 0) || (len(this.SystemErrors) > 0)
}

func (this *UserOpResult) MustClone() *UserOpResult {
	var clone UserOpResult
	util.MustJSONFromString(util.MustToJSON(this), &clone)
	return &clone
}
