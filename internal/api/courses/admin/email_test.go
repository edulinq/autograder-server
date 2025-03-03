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

		// CC
		{
			UserEmail: "course-grader",
			Message: email.Message{
				CC:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
		},

		// BCC
		{
			UserEmail: "course-grader",
			Message: email.Message{
				BCC:     []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
			},
		},

		// Body
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
				Body:    "Body",
			},
		},

		//HTML
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To:      []string{"course-student@test.edulinq.org"},
				Subject: "Subject",
				HTML:    true,
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

		// No recipients after resolving email addresses.
		{
			UserEmail: "course-grader",
			Message: email.Message{
				To:      []string{"course-student", "-course-student@test.edulinq.org"},
				Subject: "Subject",
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

		if response.Locator != "" {
			test.Errorf("Case %d: Unexpected error. Expected '%s', found '%s'.", i, testCase.Locator, response.Locator)
			continue
		}

		testMessages := email.GetTestMessages()
		if testCase.DryRun {
			if len(testMessages) != 0 {
				test.Errorf("Case %d: Email was sent when it should not have.", i)
			}

			continue
		}

		if len(testMessages) == 0 {
			test.Errorf("Case %d: Email was not sent when it should have.", i)
			continue
		}

		var responseContent EmailResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)
		testMessage := testMessages[0]

		if (len(responseContent.To) + len(responseContent.CC) + len(responseContent.BCC)) == 0 {
			test.Errorf("Case %d: Email was sent without any recipients.", i)
			continue
		}

		if testCase.Message.Body != testMessage.Body {
			test.Errorf("Case %d: Unexpected body content. Expected: '%s', actual: '%s'.", i, testCase.Message.Body, testMessage.Body)
			continue
		}

		if testCase.Message.HTML != testMessage.HTML {
			test.Errorf("Case %d: Unexpected HTML value. Expected: '%t', actual: '%t'.", i, testCase.Message.HTML, testMessage.HTML)
			continue
		}
	}
}
