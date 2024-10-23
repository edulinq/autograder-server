package api

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/util"
)

func TestDescribeFull(test *testing.T) {
	path, err := common.GetAPIDescriptionFilepath()
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

func TestDescribeDuplicateBasePaths(test *testing.T) {
	routes := []core.Route{
		&core.APIRoute{
			BaseRoute: core.BaseRoute{
				BasePath: "/api/v1/duplicate",
			},
			RequestType:  reflect.TypeOf("string"),
			ResponseType: reflect.TypeOf(123),
		},
		&core.APIRoute{
			BaseRoute: core.BaseRoute{
				BasePath: "/api/v1/duplicate",
			},
			RequestType:  reflect.TypeOf([]byte{}),
			ResponseType: reflect.TypeOf(true),
		},
	}

	description := Describe(routes)

	if len(description.Endpoints) != 1 {
		test.Errorf("Unexpected number of endpoints. Expected: '1', actual: '%d'.", len(description.Endpoints))
	}

	expected := core.EndpointDescription{
		RequestType:  "[]uint8",
		ResponseType: "bool",
	}

	if !reflect.DeepEqual(description.Endpoints["/api/v1/duplicate"], expected) {
		test.Errorf("Unexpected endpoint description. Expected '%+v', actual '%+v'.",
			expected, description.Endpoints["/api/v1/duplicate"])
	}
}
