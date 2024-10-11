package util

// A utility function for creating a string pointer for a string literal.
// Should mostly be used for testing.
func StringPointer(target string) *string {
	return &target
}

func PointerToString(target *string) string {
	if target == nil {
		return ""
	}

	return *target
}

func IntPointer(target int) *int {
	return &target
}

func PointerToInt(target *int) int {
	if target == nil {
		return 0
	}

	return *target
}
