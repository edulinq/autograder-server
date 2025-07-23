package api

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	courseUsers "github.com/edulinq/autograder/internal/api/courses/users"
	"github.com/edulinq/autograder/internal/api/users"
	"github.com/edulinq/autograder/internal/util"
)

// A test server started in `internal/api/core` will not be able to get all routes from api.GetRoutes() due to an import cycle.
// So, we test describing all API endpoints in `internal/api`.
func TestDescribeRoutesFull(test *testing.T) {
	path, err := util.GetAPIDescriptionFilepath()
	if err != nil {
		test.Fatalf("Unable to get the API description filepath: '%v'.", err)
	}

	var expectedDescriptions core.APIDescription
	err = util.JSONFromFile(path, &expectedDescriptions)
	if err != nil {
		test.Fatalf("Failed to load api.json: '%v'.", err)
	}

	actualDescriptions, err := core.DescribeRoutes(*GetRoutes())
	if err != nil {
		test.Fatalf("Failed to describe endpoints: '%v'.", err)
	}

	// Compare the JSON objects for equality to avoid any pointers or empty values.
	descriptionString := util.MustToJSON(actualDescriptions)
	var descriptions core.APIDescription
	util.MustJSONFromString(descriptionString, &descriptions)

	if reflect.DeepEqual(expectedDescriptions, descriptions) {
		return
	}

	diff := util.MustComputeTestDiff(util.MustToJSONIndent(expectedDescriptions), util.MustToJSONIndent(descriptions))
	test.Fatalf("Unexpected API Description. Diff:\n%s.", diff)
}

func TestDescribeRoutesEmptyDescription(test *testing.T) {
	descriptions, err := core.DescribeRoutes(*GetRoutes())
	if err != nil {
		test.Fatalf("Failed to describe endpoints: '%v'.", err)
	}

	for endpoint, description := range descriptions.Endpoints {
		if description.Description == "" {
			test.Errorf("Describe found an empty description. Endpoint: '%s'.", endpoint)
			continue
		}
	}
}

// Test types with conflicting names in `internal/api` to avoid cycles when importing `users.ListRequest` and `courses/users.ListRequest`.
func TestDescribeTypeConflictingNames(test *testing.T) {
	info := core.TypeInfoCache{
		TypeConversions: make(map[string]string),
	}

	// Add in the first users.ListRequest which will work.
	_, _, _, err := core.DescribeType(reflect.TypeOf((*users.ListRequest)(nil)).Elem(), true, info)
	if err != nil {
		test.Fatalf("Failed to describe type: '%v'.", err)
	}

	// Add in the second users.ListRequest which will cause a conflict.
	_, _, _, err = core.DescribeType(reflect.TypeOf((*courseUsers.ListRequest)(nil)).Elem(), true, info)
	if err == nil {
		test.Fatalf("Did not get expected error while describing types.")
	}

	expectedMessage := "Unable to get type ID due to conflicting names."
	if !strings.Contains(err.Error(), expectedMessage) {
		test.Fatalf("Did not get the expected error output. Expected substring: '%s', actual: '%s'.",
			expectedMessage, err.Error())
	}
}
