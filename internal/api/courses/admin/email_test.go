package admin

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/util"
)

func TestEmail(test *testing.T) {
	defer email.ClearTestMessages()

	testCases := []struct {
		UserEmail string
		Message   email.Message
		Locator   string
		DryRun    bool
	}{
		// No subject.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To: []string{"*"},
			},
			Locator: "627",
		},

		// No recipients.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				Subject: "Message in a Bottle",
			},
			Locator: "-628",
		},

		// Invalid permissions.
		{
			UserEmail: "course-other",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email from Course Student",
			},
			Locator: "-020",
		},
		{
			UserEmail: "course-student",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email from Course Student",
			},
			Locator: "-020",
		},

		// Invalid permissions, role escalation.
		{
			UserEmail: "server-user",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email from Course Student",
			},
			Locator: "-040",
		},
		{
			UserEmail: "server-creator",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email from Course Student",
			},
			Locator: "-040",
		},

		// Valid permissions.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email for All Course Users",
			},
		},
		{
			UserEmail: "course-admin",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email for All Course Users",
			},
		},
		{
			UserEmail: "course-owner",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email for All Course Users",
			},
		},

		// Valid permissions, role escalation.
		{
			UserEmail: "server-admin",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email for All Course Users",
			},
		},
		{
			UserEmail: "server-owner",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email for All Course Users",
			},
		},

		// Dry run.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email for All Course Users",
			},
			DryRun: true,
		},
	}

	for _, testCase := range testCases {
		email.ClearTestMessages()

		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.Message), &fields)
		fields["dry-run"] = testCase.DryRun

		response := core.SendTestAPIRequestFull(test, "courses/admin/email", fields, nil, testCase.UserEmail)
		fmt.Println(util.MustToJSONIndent(response))
		fmt.Println(util.MustToJSONIndent(email.GetTestMessages()))
	}
}
