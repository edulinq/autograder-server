package users

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

	response := core.SendTestAPIRequest(test, core.NewEndpoint(`users/password/reset`), nil)
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	messages := email.GetTestMessages()
	if len(messages) != 1 {
		test.Fatalf("Did not find the correct number of messages. Expected: 1, actual: %d.", len(messages))
	}

	to := []string{"server-admin@test.edulinq.org"}
	if !reflect.DeepEqual(to, messages[0].To) {
		test.Fatalf("Unexpected message recipients. Expected: [%s], actual: [%s].",
			strings.Join(to, ", "), strings.Join(messages[0].To, ", "))
	}

	re := regexp.MustCompile(`token: '(.*?)'`)
	matches := re.FindStringSubmatch(messages[0].Body)

	if len(matches) != 2 {
		test.Fatalf("Unexpected number of regexp matches. Expected: 2, actual: %d.", len(matches))
	}

	user, err := db.GetServerUser("server-admin@test.edulinq.org", true)
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
