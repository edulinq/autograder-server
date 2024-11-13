package ast

import (
	"testing"
)

func TestGetAPIDescriptions(test *testing.T) {
	descriptions := GetAPIDescriptions()

	test.Errorf("Found the following descriptions: '%+v'.", descriptions)
}
