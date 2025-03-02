package admin

import (
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
		PermError bool
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
			Locator:   "-020",
			PermError: true,
		},
		{
			UserEmail: "course-student",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email from Course Student",
			},
			Locator:   "-020",
			PermError: true,
		},

		// Invalid permissions, role escalation.
		{
			UserEmail: "server-user",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email from Course Student",
			},
			Locator:   "-040",
			PermError: true,
		},
		{
			UserEmail: "server-creator",
			Message: email.Message{
				To:      []string{"*"},
				Subject: "Email from Course Student",
			},
			Locator:   "-040",
			PermError: true,
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

	for i, testCase := range testCases {
		email.ClearTestMessages()

		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.Message), &fields)
		fields["dry-run"] = testCase.DryRun

		response := core.SendTestAPIRequestFull(test, "courses/admin/email", fields, nil, testCase.UserEmail)

		if !response.Success {
			if testCase.PermError {
				if response.Locator != testCase.Locator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, testCase.Locator, response.Locator)
				}
			} else if testCase.Message.Subject == "" {
				continue
			} else if (len(testCase.Message.To) + len(testCase.Message.CC) + len(testCase.Message.BCC)) == 0 {
				continue
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.PermError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		if testCase.DryRun {
			if email.GetTestMessages() != nil {
				test.Errorf("Case %d: Emails were sent when they should not have.", i)
			}
		}
	}
}
