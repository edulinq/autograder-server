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

func (this *ModelError) MustClone() *ModelError {
	clone := ModelError{
		Locator:         this.Locator,
		InternalMessage: this.InternalMessage,
		ExternalMessage: this.ExternalMessage,
	}

	return &clone
}

func DereferenceModelErrors(errors []*ModelError) []ModelError {
	result := make([]ModelError, len(errors))

	for i, err := range errors {
		if err != nil {
			result[i] = ModelError{
				Locator:         err.Locator,
				InternalMessage: err.InternalMessage,
				ExternalMessage: err.ExternalMessage,
			}
		}
	}

	return result
}

func (this ModelError) Equals(other ModelError) bool {
	if this.Locator != other.Locator {
		return false
	}

	if this.InternalMessage != other.InternalMessage {
		return false
	}

	if this.ExternalMessage != other.ExternalMessage {
		return false
	}

	return true
}

func ModelErrorSlicesEquals(a []*ModelError, b []*ModelError) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] == b[i] {
			continue
		}

		if a[i] == nil {
			return false
		}

		if b[i] == nil {
			return false
		}

		if !a[i].Equals(*b[i]) {
			return false
		}
	}

	return true
}
