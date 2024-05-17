package util

import (
	"testing"
)

type Unmarshalable struct {
	SomeFunc func() error
}

func (this *Unmarshalable) String() string {
	return BaseString(this)
}

// Test that unmarshalable values do not infinite loop
// (since we may try to print them when writing errors).
func TestUnmarshalableStruct(test *testing.T) {
	value := &Unmarshalable{}

	_, err := ToJSON(value)
	if err == nil {
		test.Errorf("Unmarshalable object did not return error.")
	}
}
