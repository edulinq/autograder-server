package model

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

// A general representation of the result of operations that modify a user in any way (add, remove, enroll, drop, etc).
type BaseUserOpResult struct {
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
}

// A general representation of the result of operations that modify a user in any way (add, remove, enroll, drop, etc).
// All user-facing functions (essentially non-db functions) should return an instance or collection of these objects.
type UserOpResult struct {
	// Embed BaseUserOpResult for basic fields, see above.
	BaseUserOpResult

	// The following error occurred during this operation because of the provided data,
	// i.e., they are caused by the calling user.
	// All error messages should be safe for users.
	ValidationError *LocatableError `json:"validation-error,omitempty"`

	// The following error occurred during this operation, but not because of the provided data,
	// i.e., they are the system's fault.
	// These errors are not guaranteed to be safe for users,
	// and the calling code should decide how they should be managed.
	SystemError *LocatableError `json:"system-error,omitempty"`

	// The following error occurred during this operation, but not because of the provided data,
	// i.e., the system was unable to communicate the results.
	// These errors are not guaranteed to be safe for users,
	// and the calling code should decide how they should be managed.
	CommunicationError *LocatableError `json:"communication-error,omitempty"`

	// The following cleartext password was generated during this operation.
	// Care should be taken to not expose this field.
	CleartextPassword string `json:"cleartext-password,omitempty"`
}

// A user safe representation of the UserOpResult struct.
// Notably all errors will be converted to responses and the cleartext password field is removed.
// For descriptions of shared fields, see UserOpResult above.
type ExternalUserOpResult struct {
	// Embed BaseUserOpResult for basic fields, see above.
	BaseUserOpResult

	// A user safe representation of a validation error, which will not include a locator.
	ValidationError *ExternalLocatableError `json:"validation-error,omitempty"`

	// A user safe representation of a system error.
	SystemError *ExternalLocatableError `json:"system-error,omitempty"`

	// A user safe representation of a communication error.
	CommunicationError *ExternalLocatableError `json:"communication-error,omitempty"`
}

// A struct containing counts summarizing results.
// Each value will get +1 for a result that has a matching non-empty value.
// This means that enrollment/errors will only get +1 regardless of the number of members over 0.
// Mainly useful for testing.
type UserOpResultsCounts struct {
	Total              int
	Added              int
	Modified           int
	Removed            int
	Skipped            int
	NotExists          int
	Emailed            int
	Enrolled           int
	Dropped            int
	ValidationError    int
	SystemError        int
	CommunicationError int
	CleartextPassword  int
}

func NewUserOpResultValidationError(locator string, email string, err error) *UserOpResult {
	return &UserOpResult{
		BaseUserOpResult: BaseUserOpResult{
			Email: email,
		},
		ValidationError: NewLocatableError(locator, true, err.Error(), fmt.Sprintf("You have insufficient permissions for the requested operation.")),
	}
}

func NewUserOpResultSystemError(locator string, email string, err error) *UserOpResult {
	return &UserOpResult{
		BaseUserOpResult: BaseUserOpResult{
			Email: email,
		},
		SystemError: NewLocatableError(locator, false, err.Error(),
			fmt.Sprintf("The server failed to process your request. Please contact an administrator with this ID '%s'.", locator)),
	}
}

func (this *UserOpResult) HasErrors() bool {
	return (this.ValidationError != nil) || (this.SystemError != nil) || (this.CommunicationError != nil)
}

func (this *UserOpResult) MustClone() *UserOpResult {
	var clone UserOpResult
	util.MustJSONFromString(util.MustToJSON(this), &clone)

	return &clone
}

func (this *UserOpResult) ToExternalResult() *ExternalUserOpResult {
	var externalValError, externalSysError, externalCommError *ExternalLocatableError

	if this.ValidationError != nil {
		externalValError = this.ValidationError.ToExternalError()
	}

	if this.SystemError != nil {
		externalSysError = this.SystemError.ToExternalError()
	}

	if this.CommunicationError != nil {
		externalCommError = this.CommunicationError.ToExternalError()
	}

	return &ExternalUserOpResult{
		BaseUserOpResult:   this.BaseUserOpResult,
		ValidationError:    externalValError,
		SystemError:        externalSysError,
		CommunicationError: externalCommError,
	}
}

func CompareUserOpResultPointer(a *UserOpResult, b *UserOpResult) int {
	if a == b {
		return 0
	}

	if a == nil {
		return 1
	}

	if b == nil {
		return -1
	}

	return CompareUserOpResult(*a, *b)
}

func CompareUserOpResult(a UserOpResult, b UserOpResult) int {
	return strings.Compare(a.Email, b.Email)
}

func CompareExternalUserOpResultPointer(a *ExternalUserOpResult, b *ExternalUserOpResult) int {
	if a == b {
		return 0
	}

	if a == nil {
		return 1
	}

	if b == nil {
		return -1
	}

	return CompareExternalUserOpResult(*a, *b)
}

func CompareExternalUserOpResult(a ExternalUserOpResult, b ExternalUserOpResult) int {
	return strings.Compare(a.Email, b.Email)
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
		counts.ValidationError += boolToInt(result.ValidationError != nil)
		counts.SystemError += boolToInt(result.SystemError != nil)
		counts.CommunicationError += boolToInt(result.CommunicationError != nil)
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
