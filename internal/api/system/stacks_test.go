package system

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestStacksBase(test *testing.T) {
	response := core.SendTestAPIRequestFull(test, `system/stacks`, nil, nil, "server-admin")
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent StacksResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	// Minimum stacks: test runner, this test.
	minStacks := 2
	if len(responseContent.Stacks) < minStacks {
		test.Fatalf("Only found %d stacks, expecting at least %d.", len(responseContent.Stacks), minStacks)
	}
}
