package model

// A general representation of locatable errors that can be generated.
type LocatableError struct {
	// The locator for the error which is not exported.
	Locator string

	// A flag for authentication errors so we know to hide locators before responding.
	AuthError bool

	// The internal message for the error which is not exported.
	InternalMessage string

	// The external message of the error which MUST be user friendly.
	ExternalMessage string
}

type LocatableErrorResponse struct {
	Locator string `json:"locator"`
	Message string `json:"message"`
}

func NewLocatableError(locator string, authError bool, internalMessage string, externalMessage string) *LocatableError {
	return &LocatableError{
		Locator:         locator,
		AuthError:       authError,
		InternalMessage: internalMessage,
		ExternalMessage: externalMessage,
	}
}

func (this *LocatableError) ToResponse() *LocatableErrorResponse {
	// Remove the locator for authentication errors.
	locator := this.Locator
	if this.AuthError {
		locator = ""
	}

	return &LocatableErrorResponse{
		Locator: locator,
		Message: this.ExternalMessage,
	}
}
