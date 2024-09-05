package model

import (
	"fmt"

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
	ValidationErrors []*ModelError `json:"validation-errors,omitempty"`

	// The following error occurred during this operation, but not because of the provided data,
	// i.e., they are the system's fault.
	// These errors are not guarenteed to be safe for users,
	// and the calling code should decide how they should be managed.
	SystemErrors []*ModelError `json:"system-errors,omitempty"`

	// The following cleartext password was generated during this operation.
	// Care should be taken to not expose their field.
	CleartextPassword string `json:"cleartext-password,omitempty"`
}

// A struct containg counts summarizing results.
// Each value will get +1 for a result that has a matching non-empty value.
// This means that enrollment/errors will only get +1 regardless of the number of members over 0.
// Mainly useful for testing.
type UserOpResultsCounts struct {
	Total             int
	Added             int
	Modified          int
	Removed           int
	Skipped           int
	NotExists         int
	Emailed           int
	Enrolled          int
	Dropped           int
	ValidationErrors  int
	SystemErrors      int
	CleartextPassword int
}

func NewUserOpResultValidationError(locator string, email string, err error) *UserOpResult {
	return &UserOpResult{
		Email: email,
		ValidationErrors: []*ModelError{
			NewModelError(locator, err.Error(), fmt.Sprintf("You have insufficient permissions for the requested operation.")),
		},
	}
}

func NewUserOpResultSystemError(locator string, email string, err error) *UserOpResult {
	return &UserOpResult{
		Email: email,
		SystemErrors: []*ModelError{
			NewModelError(locator, err.Error(),
				fmt.Sprintf("The server failed to process your request. Please contact an administrator with this ID '%s'.", locator)),
		},
	}
}

func (this *UserOpResult) HasErrors() bool {
	return (len(this.ValidationErrors) > 0) || (len(this.SystemErrors) > 0)
}

func (this *UserOpResult) MustClone() *UserOpResult {
	var clone UserOpResult
	util.MustJSONFromString(util.MustToJSON(this), &clone)

	// Clone the ModelErrors to preservee non exported fields.
	clone.ValidationErrors = make([]*ModelError, 0, len(this.ValidationErrors))
	for i := range this.ValidationErrors {
		clone.ValidationErrors = append(clone.ValidationErrors, this.ValidationErrors[i].MustClone())
	}

	clone.SystemErrors = make([]*ModelError, 0, len(this.SystemErrors))
	for i := range this.SystemErrors {
		clone.SystemErrors = append(clone.SystemErrors, this.SystemErrors[i].MustClone())
	}

	return &clone
}

func GetUserOpResultsCounts(results []*UserOpResult) *UserOpResultsCounts {
	counts := UserOpResultsCounts{}

	for _, result := range results {
		counts.Total++

		counts.Added += boolToInt(result.Added)
		counts.Modified += boolToInt(result.Modified)
		counts.Removed += boolToInt(result.Removed)
		counts.Skipped += boolToInt(result.Skipped)
		counts.NotExists += boolToInt(result.NotExists)
		counts.Emailed += boolToInt(result.Emailed)
		counts.Enrolled += boolToInt(len(result.Enrolled) > 0)
		counts.Dropped += boolToInt(len(result.Dropped) > 0)
		counts.ValidationErrors += boolToInt(len(result.ValidationErrors) > 0)
		counts.SystemErrors += boolToInt(len(result.SystemErrors) > 0)
		counts.CleartextPassword += boolToInt(result.CleartextPassword != "")
	}

	return &counts
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}
