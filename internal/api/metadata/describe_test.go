package metadata

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

func TestMetadataDescribe(test *testing.T) {
	defer db.ResetForTesting()

	// Cache a dummy APIDescription for testing.
	description := &core.APIDescription{
		Endpoints: map[string]core.EndpointDescription{
			"metadata/describe": core.EndpointDescription{},
		},
	}

	testCases := []struct {
		ForceUpdate bool
		Routes      *[]core.Route
		Description *core.APIDescription
		Locator     string
	}{
		// Use cached description.
		{
			false,
			nil,
			description,
			"",
		},

		// Force update to describe one route.
		{
			true,
			&[]core.Route{core.MustNewAPIRoute(`metadata/describe`, HandleDescribe)},
			&core.APIDescription{
				Endpoints: map[string]core.EndpointDescription{
					"metadata/describe": core.EndpointDescription{
						Description: "Describe all endpoints on the server.",
						Input: []core.FieldDescription{
							core.FieldDescription{
								Name: "force-update",
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
						},
					},
				},
			},
			"",
		},

		// Force update without any routes cached in `internal/api/core`.
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
		db.ResetForTesting()

		oldDescription := mustGetAPIDescription()
		mustSetAPIDescription(description)
		defer mustSetAPIDescription(oldDescription)

		oldRoutes := core.GetAPIRoutes()
		core.SetAPIRoutes(testCase.Routes)
		defer core.SetAPIRoutes(oldRoutes)

		fields := map[string]any{"force-update": testCase.ForceUpdate}

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

		expected := DescribeResponse{*testCase.Description}
		if !reflect.DeepEqual(expected, responseContent) {
			test.Fatalf("Unexpected API description. Expected: '%s', actual: '%s'.",
				util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent))
		}
	}
}

func mustSetAPIDescription(apiDescription *core.APIDescription) {
	err := core.SetAPIDescription(apiDescription)
	if err != nil {
		log.Fatal("Unable to cache API descriptions.", err)
	}
}

func mustGetAPIDescription() *core.APIDescription {
	apiDescription, err := core.GetAPIDescription()
	if err != nil {
		log.Fatal("Unable to get cached API description.", err)
	}

	return apiDescription
}
