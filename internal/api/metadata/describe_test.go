package metadata

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestMetadataDescribe(test *testing.T) {
	description := &core.APIDescription{
		Endpoints: map[string]core.EndpointDescription{
			"metadata/describe": core.EndpointDescription{
				Description: "Describe all endpoints on the server.",
				Input: []core.FieldDescription{
					core.FieldDescription{
						Name: "force-compute",
						Type: "bool",
					},
				},
				Output: []core.FieldDescription{
					core.FieldDescription{
						Name: "endpoints",
						Type: "map[string]core.EndpointDescription",
					},
					core.FieldDescription{
						Name: "types",
						Type: "map[string]core.TypeDescription",
					},
				},
			},
		},
		Types: map[string]core.TypeDescription{
			"core.APIDescription": core.TypeDescription{
				Category: "struct",
				Fields: []core.FieldDescription{
					{
						Name: "endpoints",
						Type: "map[string]core.EndpointDescription",
					},
					{
						Name: "types",
						Type: "map[string]core.TypeDescription",
					},
				},
			},
			"core.EndpointDescription": core.TypeDescription{
				Category: "struct",
				Fields: []core.FieldDescription{
					{
						Name: "description",
						Type: "string",
					},
					{
						Name: "input",
						Type: "[]core.FieldDescription",
					},
					{
						Name: "output",
						Type: "[]core.FieldDescription",
					},
				},
			},
			"core.FieldDescription": core.TypeDescription{
				Category: "struct",
				Fields: []core.FieldDescription{
					{
						Name: "name",
						Type: "string",
					},
					{
						Name: "type",
						Type: "string",
					},
				},
			},
			"core.TypeDescription": core.TypeDescription{
				Category: "struct",
				Fields: []core.FieldDescription{
					{
						Name: "alias-type",
						Type: "string",
					},
					{
						Name: "category",
						Type: "string",
					},
					{
						Name: "description",
						Type: "string",
					},
					{
						Name: "element-type",
						Type: "string",
					},
					{
						Name: "fields",
						Type: "[]core.FieldDescription",
					},
					{
						Name: "key-type",
						Type: "string",
					},
					{
						Name: "value-type",
						Type: "string",
					},
				},
			},
		},
	}

	testCases := []struct {
		ForceCompute bool
		Routes       *[]core.Route
		Description  *core.APIDescription
		Locator      string
	}{
		// Use cached description.
		{
			false,
			nil,
			description,
			"",
		},

		// Force compute to describe one route.
		{
			true,
			&[]core.Route{core.MustNewAPIRoute(`metadata/describe`, HandleDescribe)},
			description,
			"",
		},

		// Force compute without any routes cached in `internal/api/core`.
		{
			true,
			nil,
			&core.APIDescription{},
			"-501",
		},
		{
			true,
			&[]core.Route{},
			&core.APIDescription{},
			"-501",
		},
	}

	for i, testCase := range testCases {
		oldRoutes := core.GetAPIRoutes()
		core.SetAPIRoutes(testCase.Routes)
		defer core.SetAPIRoutes(oldRoutes)

		fields := map[string]any{"force-compute": testCase.ForceCompute}

		response := core.SendTestAPIRequestFull(test, `metadata/describe`, fields, nil, "course-student")
		if !response.Success {
			if testCase.Locator != response.Locator {
				test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
					i, testCase.Locator, response.Locator)
			}

			continue
		}

		if testCase.Locator != "" {
			test.Errorf("Case %d: Did not get an expected error. Expected: '%s'", i, testCase.Locator)
			continue
		}

		var responseContent DescribeResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		expected := DescribeResponse{testCase.Description}
		if testCase.ForceCompute {
			if !reflect.DeepEqual(expected, responseContent) {
				test.Errorf("Case %d: Unexpected API description. Expected: '%s', actual: '%s'.",
					i, util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent))
				continue
			}
		} else {
			if !subsetEqualityCheck(test, i, testCase.Description, responseContent.APIDescription) {
				continue
			}
		}
	}
}

// If the API description is not computed, the description comes from `resources/api.json`.
// As that resource constantly evolves and is tested for correctness in `internal/api/describe_test.go`,
// only check that a subset of the API description is correctly returned by `metadata/describe`.
func subsetEqualityCheck(test *testing.T, testNum int, expected *core.APIDescription, actual *core.APIDescription) bool {
	for endpoint, expectedDescription := range expected.Endpoints {
		actualDescription, ok := actual.Endpoints[endpoint]
		if !ok {
			test.Errorf("Case %d: Could not find an expected description for endpoint '%s'.", testNum, endpoint)
			return false
		}

		if !reflect.DeepEqual(expectedDescription, actualDescription) {
			test.Errorf("Case %d: Unexpected description for endpoint '%s'. Expected: '%s', actual: '%s'.",
				testNum, endpoint, util.MustToJSONIndent(expectedDescription), util.MustToJSONIndent(actualDescription))
			return false
		}
	}

	for customType, expectedDescription := range expected.Types {
		actualDescription, ok := actual.Types[customType]
		if !ok {
			test.Errorf("Case %d: Could not find an expected description for type '%s'.", testNum, customType)
			return false
		}

		if !reflect.DeepEqual(expectedDescription, actualDescription) {
			test.Errorf("Case %d: Unexpected description for type '%s'. Expected: '%s', actual: '%s'.",
				testNum, customType, util.MustToJSONIndent(expectedDescription), util.MustToJSONIndent(actualDescription))
			return false
		}
	}

	return true
}
