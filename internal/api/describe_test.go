package api

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestDescribeFull(test *testing.T) {
	path, err := util.GetAPIDescriptionFilepath()
	if err != nil {
		test.Fatalf("Unable to get the API description filepath: '%v'.", err)
	}

	var expectedDescription core.APIDescription
	err = util.JSONFromFile(path, &expectedDescription)
	if err != nil {
		test.Fatalf("Failed to load api.json: '%v'.", err)
	}

	fullDescription := Describe(*GetRoutes())

	if !reflect.DeepEqual(&expectedDescription, fullDescription) {
		test.Fatalf("Unexpected API Descriptions. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expectedDescription), util.MustToJSONIndent(fullDescription))
	}
}

func TestDescribeEmptyRoutes(test *testing.T) {
	routes := []core.Route{}
	description := Describe(routes)

	if len(description.Endpoints) != 0 {
		test.Errorf("Unexpected number of endpoints. Expected: '0', actual: '%d'.", len(description.Endpoints))
	}
}
