package model

// A general representation of errors that have a definite source location.
type LocatableError struct {
	// The locator for the error which is not exported.
	Locator string

	// A flag for knowing when to hide locators before responding.
	HideLocator bool

	// The internal message for the error which is not exported.
	InternalMessage string

	// The external message of the error which MUST be user friendly.
	ExternalMessage string
}

// A user safe version of locatable errors.
// All LocatableErrors must be converted to ExternalLocatableErrors
// if it is to be given to a user.
type ExternalLocatableError struct {
	Locator string `json:"locator"`
	Message string `json:"message"`
}

func NewLocatableError(locator string, hideLocator bool, internalMessage string, externalMessage string) *LocatableError {
	return &LocatableError{
		Locator:         locator,
		HideLocator:     hideLocator,
		InternalMessage: internalMessage,
		ExternalMessage: externalMessage,
	}
}

func (this *LocatableError) ToExternalError() *ExternalLocatableError {
	// Hide the locator if necessary.
	locator := this.Locator
	if this.HideLocator {
		locator = ""
	}

	return &ExternalLocatableError{
		Locator: locator,
		Message: this.ExternalMessage,
	}
}
