package util

import (
	"reflect"
)

// Check for nil, even if the target is an interface (which requires an additional check).
func IsNil(value any) bool {
	if value == nil {
		return true
	}

	// The value type may still be a pointer (or other nil-able) types.
	// reflect.IsNil() cannot be called on non-nil-able types.
	switch reflect.TypeOf(value).Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflect.ValueOf(value).IsNil()
	default:
		return false
	}
}
