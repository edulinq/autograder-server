package users

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestTokensDelete(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	user, err := db.GetServerUser("admin@test.com", true)
	if err != nil {
		test.Fatalf("Failed to get initial user: '%v'.", err)
	}

	initialTokenCount := len(user.Tokens)
	if initialTokenCount == 0 {
		test.Fatalf("Test user has no tokens.")
	}

	args := map[string]any{
		"token-id": user.Tokens[0].ID,
	}

	response := core.SendTestAPIRequest(test, core.NewEndpoint("server/users/tokens/delete"), args)
	if !response.Success {
		test.Fatalf("Response not successful: '%s'.", util.MustToJSONIndent(response))
	}

	user, err = db.GetServerUser("admin@test.com", true)
	if err != nil {
		test.Fatalf("Failed to get saved user: '%v'.", err)
	}

	newTokenCount := len(user.Tokens)

	if newTokenCount != (initialTokenCount - 1) {
		test.Fatalf("Incorrect token count. Expected: %d, Found: %d.", (initialTokenCount - 1), newTokenCount)
	}
}
