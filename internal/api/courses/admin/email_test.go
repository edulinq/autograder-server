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
	}{
		// Valid permissions.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
		},

		// Valid permissions, role escalation.
		{
			UserEmail: "server-admin",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
		},
		{
			UserEmail: "server-owner",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
		},

		// Ensure recipients check occurs after resolution.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To:      []string{"course-student", "course-student@test.edulinq.org"},
				Subject: "Subject",
			},
		},

		// Dry run.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
			DryRun: true,
		},

		// Invalid permissions.
		{
			UserEmail: "course-student",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
			Locator: "-020",
		},

		// Invalid permissions, role escalation.
		{
			UserEmail: "server-user",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
			Locator: "-040",
		},
		{
			UserEmail: "server-creator",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
			Locator: "-040",
		},

		// No subject.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To: []string{"course-student@test.edulinq.org"},
			},
			Locator: "-627",
		},

		// No recipients.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				Subject: "Message in a Bottle",
			},
			Locator: "-629",
		},
	}

	for i, testCase := range testCases {
		email.ClearTestMessages()

		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.Message), &fields)
		fields["dry-run"] = testCase.DryRun

		response := core.SendTestAPIRequestFull(test, "courses/admin/email", fields, nil, testCase.UserEmail)

		if !response.Success {
			if response.Locator != testCase.Locator {
				test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
					i, testCase.Locator, response.Locator)
			}

			continue
		}

		if testCase.DryRun {
			if len(email.GetTestMessages()) != 0 {
				test.Errorf("Case %d: Email was sent when it should not have.", i)
			}

			continue
		}
	}
}
