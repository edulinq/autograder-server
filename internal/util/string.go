package util

import (
	"fmt"
	"slices"
	"strings"
)

func BaseString(obj any) string {
	json, err := ToJSON(obj)
	if err != nil {
		// Explicitly use Go-Syntax (%#v) to avoid loops with overwritten String() methods.
		return fmt.Sprintf("%#v", obj)
	}

	return json
}

func JoinStrings(delim string, parts ...string) string {
	return strings.Join(parts, delim)
}

func GetStringWithDefault(primary string, fallback string) string {
	if primary != "" {
		return primary
	}

	return fallback
}

func SortStringsCopy(input []string) []string {
	output := input[:]
	slices.Sort(output)
	return output
}

func StringsEqualsIgnoreOrdering(a []string, b []string) bool {
	aSorted := SortStringsCopy(a)
	bSorted := SortStringsCopy(b)
	return slices.Equal(aSorted, bSorted)
}

func StringsCompareIgnoreOrdering(a []string, b []string) int {
	aSorted := SortStringsCopy(a)
	bSorted := SortStringsCopy(b)
	return slices.Compare(aSorted, bSorted)
}

func StringContainedInSlice(str string, slice []string) bool {
	for _, value := range slice {
		if str == value {
			return true
		}
	}

	return false
}
