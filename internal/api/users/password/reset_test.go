package password

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/util"
)

func TestPassReset(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	email.ClearTestMessages()

	// Reset the authenticated user's own password (no target-user-email).
	response := core.SendTestAPIRequest(test, `users/password/reset`, nil)
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	messages := email.GetTestMessages()
	if len(messages) != 1 {
		test.Fatalf("Did not find the correct number of messages. Expected: 1, actual: %d.", len(messages))
	}

	to := []string{"course-admin@test.edulinq.org"}
	if !reflect.DeepEqual(to, messages[0].To) {
		test.Fatalf("Unexpected message recipients. Expected: [%s], actual: [%s].",
			strings.Join(to, ", "), strings.Join(messages[0].To, ", "))
	}

	re := regexp.MustCompile(`token: '(.*?)'`)
	matches := re.FindStringSubmatch(messages[0].Body)

	if len(matches) != 2 {
		test.Fatalf("Unexpected number of regexp matches. Expected: 2, actual: %d.", len(matches))
	}

	user, err := db.GetServerUser("course-admin@test.edulinq.org")
	if err != nil {
		test.Fatalf("Failed to get saved user: '%v'.", err)
	}

	success, err := user.Auth(util.Sha256HexFromString(matches[1]))
	if err != nil {
		test.Fatalf("Failed to auth user: '%v'.", err)
	}

	if !success {
		test.Fatalf("The new password fails to auth.")
	}
}

// TestPassResetTargetAdmin verifies that a server admin can reset another user's password.
func TestPassResetTargetAdmin(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	email.ClearTestMessages()

	// server-admin resets course-student's password.
	fields := map[string]any{
		"target-user-email": "course-student@test.edulinq.org",
	}

	response := core.SendTestAPIRequestFull(test, `users/password/reset`, fields, nil, "server-admin")
	if !response.Success {
		test.Fatalf("Admin reset: response is not a success: '%v'.", response)
	}

	messages := email.GetTestMessages()
	if len(messages) != 1 {
		test.Fatalf("Admin reset: expected 1 email, got %d.", len(messages))
	}

	if !reflect.DeepEqual([]string{"course-student@test.edulinq.org"}, messages[0].To) {
		test.Fatalf("Admin reset: email sent to wrong recipient: '%s'.", strings.Join(messages[0].To, ", "))
	}
}

// TestPassResetTargetNonAdmin verifies that a non-admin cannot reset another user's password.
func TestPassResetTargetNonAdmin(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	email.ClearTestMessages()

	fields := map[string]any{
		"target-user-email": "course-admin@test.edulinq.org",
	}

	// course-student tries to reset course-admin's password.
	response := core.SendTestAPIRequestFull(test, `users/password/reset`, fields, nil, "course-student")
	if response.Success {
		test.Fatalf("Non-admin reset of another user's password should have failed.")
	}

	// No email should have been sent.
	messages := email.GetTestMessages()
	if len(messages) != 0 {
		test.Fatalf("Non-admin reset: expected 0 emails, got %d.", len(messages))
	}
}
