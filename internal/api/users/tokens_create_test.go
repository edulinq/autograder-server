package users

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/util"
)

func TestTokensCreate(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	user, err := db.GetServerUser("course-admin@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get initial user: '%v'.", err)
	}

	initialTokenCount := len(user.Tokens)

	response := core.SendTestAPIRequest(test, `users/tokens/create`, nil)
	if !response.Success {
		test.Fatalf("Response not successful: '%s'.", util.MustToJSONIndent(response))
	}

	user, err = db.GetServerUser("course-admin@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get saved user: '%v'.", err)
	}

	newTokenCount := len(user.Tokens)

	if newTokenCount != (initialTokenCount + 1) {
		test.Fatalf("Incorrect token count. Expected: %d, Found: %d.", (initialTokenCount + 1), newTokenCount)
	}
}
