package metadata

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestMetadataDescribe(test *testing.T) {
	// Cache a dummy APIDescription for testing.
	description := core.APIDescription{
		Endpoints: map[string]core.EndpointDescription{
			"metadata/describe": core.EndpointDescription{},
		},
	}

	oldDescription := core.GetAPIDescription()
	core.SetAPIDescription(description)
	defer core.SetAPIDescription(oldDescription)

	response := core.SendTestAPIRequest(test, `metadata/describe`, nil)
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent DescribeResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	expected := DescribeResponse{description}
	if !reflect.DeepEqual(expected, responseContent) {
		test.Fatalf("Unexpected API description. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent))
	}
}
