package util

// Returns a new map where the original keys and values are swapped.
// For map values that are not unique, only one pair using that value as a key will exist.
func MapReverse[K comparable, V comparable](original map[K]V) map[V]K {
	rev := make(map[V]K, len(original))
	for key, value := range original {
		rev[value] = key
	}

	return rev
}
