package model

import (
	"fmt"

	"github.com/edulinq/autograder/internal/util"
)

// A general representation of the result of operations that modify a user in any way (add, remove, enroll, drop, etc).
// All user-facing functions (essentially non-db functions) should return an instance or collection of these objects.
type UserOpResult struct {
	// The email/id of the target user.
	Email string

	// The user was added to the server.
	Added bool

	// The user existed before this operation and was edited (including enrollment changes).
	Modified bool

	// The user existed before this operation and was removed.
	Removed bool

	// The user was skipped (often because they already exist).
	Skipped bool

	// The user did not exist before this operation and does not exist after.
	// This may also be an error depending on the semantics of the operation.
	NotExists bool

	// The user was emailed during the course of this operation.
	// This is more than just GetEmail() was called, an actual email was sent
	// (or would have been sent if this operation was during a dry-run).
	Emailed bool

	// The user was enrolled in the following courses (by id).
	Enrolled []string

	// The user was removed from the following courses (by id).
	Dropped []string

	// The following error occurred during this operation because of the provided data,
	// i.e., they are caused by the calling user.
	// All error messages should be safe for users.
	ValidationError *LocatableError

	// The following error occurred during this operation, but not because of the provided data,
	// i.e., they are the system's fault.
	// These errors are not guarenteed to be safe for users,
	// and the calling code should decide how they should be managed.
	SystemError *LocatableError

	// The following error occurred during this operation, but not because of the provided data,
	// i.e., the system was unable to communicate the results.
	// These errors are not guarenteed to be safe for users,
	// and the calling code should decide how they should be managed.
	CommunicationError *LocatableError

	// The following cleartext password was generated during this operation.
	// Care should be taken to not expose their field.
	CleartextPassword string
}

// A user safe representation of the UserOpResult struct.
// Notably all errors will be converted to responses and the cleartext password field is removed.
// For descriptions of shared fields, see UserOpResult above.
type UserOpResponse struct {
	Email string `json:"email"`

	Added bool `json:"added,omitempty"`

	Modified bool `json:"modified,omitempty"`

	Removed bool `json:"removed,omitempty"`

	Skipped bool `json:"skipped,omitempty"`

	NotExists bool `json:"not-exists,omitempty"`

	Emailed bool `json:"emailed,omitempty"`

	Enrolled []string `json:"enrolled,omitempty"`

	Dropped []string `json:"dropped,omitempty"`

	// A user safe representation of a validation error, which will not include a locator.
	ValidationError *LocatableErrorResponse `json:"validation-error,omitempty"`

	// A user safe representation of a system error.
	SystemError *LocatableErrorResponse `json:"system-error,omitempty"`

	// A user safe representation of a communication error.
	CommunicationError *LocatableErrorResponse `json:"communication-error,omitempty"`
}

// A struct containg counts summarizing results.
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
		Email:           email,
		ValidationError: NewLocatableError(locator, true, err.Error(), fmt.Sprintf("You have insufficient permissions for the requested operation.")),
	}
}

func NewUserOpResultSystemError(locator string, email string, err error) *UserOpResult {
	return &UserOpResult{
		Email: email,
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

func (this *UserOpResult) ToResponse() *UserOpResponse {
	var valErrorResponse, sysErrorResponse, commErrorResponse *LocatableErrorResponse

	if this.ValidationError != nil {
		valErrorResponse = this.ValidationError.ToResponse()
	}

	if this.SystemError != nil {
		sysErrorResponse = this.SystemError.ToResponse()
	}

	if this.CommunicationError != nil {
		commErrorResponse = this.CommunicationError.ToResponse()
	}

	return &UserOpResponse{
		Email:              this.Email,
		Added:              this.Added,
		Modified:           this.Modified,
		Removed:            this.Removed,
		Skipped:            this.Skipped,
		NotExists:          this.NotExists,
		Emailed:            this.Emailed,
		Enrolled:           this.Enrolled,
		Dropped:            this.Dropped,
		ValidationError:    valErrorResponse,
		SystemError:        sysErrorResponse,
		CommunicationError: commErrorResponse,
	}
}

func CompareUserOpResponsePointer(a *UserOpResponse, b *UserOpResponse) int {
    if a == b {
        return 0
    }

    if a == nil {
        return 1
    }

    if b == nil {
        return -1
    }

    return CompareUserOpResponse(*a, *b)
}

func CompareUserOpResponse(a UserOpResponse, b UserOpResponse) int {
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
