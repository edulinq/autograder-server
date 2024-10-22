package metadata

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestMetadataDescribe(test *testing.T) {
	response := core.SendTestAPIRequest(test, `metadata/describe`, nil)
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent DescribeResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	expected := DescribeResponse{*core.Describe(core.GetServerRoutes())}
	if !reflect.DeepEqual(expected, responseContent) {
		test.Fatalf("Unexpected API description. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent))
	}
}
