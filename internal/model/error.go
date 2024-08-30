package model

// A general reporesentation of errors that can be generated at the model level.
type ModelError struct {
	// The locator for the error which is not exported.
	Locator string `json:"-"`

	// The internal message for the error which is not exported.
	InternalMessage string `json:"-"`

	// The external message of the error which MUST be user friendly.
	ExternalMessage string `json:"external-message"`
}

func NewModelError(locator string, internalMessage string, externalMessage string) *ModelError {
	return &ModelError{
		Locator:         locator,
		InternalMessage: internalMessage,
		ExternalMessage: externalMessage,
	}
}
