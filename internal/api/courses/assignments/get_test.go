package assignments

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestGet(test *testing.T) {
	assignment := db.MustGetTestAssignment()
	expected := core.NewAssignmentInfo(assignment)

	response := core.SendTestAPIRequest(test, core.NewEndpoint(`courses/assignments/get`), nil)
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent GetResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if !reflect.DeepEqual(expected, responseContent.Assignment) {
		test.Fatalf("Unexpected result. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(responseContent.Assignment))
	}
}
