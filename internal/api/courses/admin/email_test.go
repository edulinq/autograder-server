package admin

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestEmail(test *testing.T) {
	defer email.ClearTestMessages()

	testCases := []struct {
		model.CourseMessageRecipients
		email.MessageContent

		UserEmail string
		Locator   string
		DryRun    bool
	}{
		// Valid permissions.
		{
			UserEmail: "course-grader",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
		},

		// Valid permissions, role escalation.
		{
			UserEmail: "server-admin",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
		},
		{
			UserEmail: "server-owner",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
		},

		// CC
		{
			UserEmail: "course-grader",
			CourseMessageRecipients: model.CourseMessageRecipients{
				CC: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
		},

		// BCC
		{
			UserEmail: "course-grader",
			CourseMessageRecipients: model.CourseMessageRecipients{
				BCC: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
		},

		// Body
		{
			UserEmail: "course-grader",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
				Body:    "Body",
			},
		},

		//HTML
		{
			UserEmail: "course-grader",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
				HTML:    true,
			},
		},

		// Dry run.
		{
			UserEmail: "course-grader",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
			DryRun: true,
		},

		// Invalid permissions.
		{
			UserEmail: "course-student",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-020",
		},

		// Invalid permissions, role escalation.
		{
			UserEmail: "server-user",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-040",
		},
		{
			UserEmail: "server-creator",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-040",
		},

		// No subject.
		{
			UserEmail: "course-grader",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Locator: "-627",
		},

		// No recipients.
		{
			UserEmail: "course-grader",
			MessageContent: email.MessageContent{
				Subject: "Message in a Bottle",
			},
			Locator: "-628",
		},

		// No recipients after resolving email addresses.
		{
			UserEmail: "course-grader",
			CourseMessageRecipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student", "-course-student@test.edulinq.org"},
			},
			MessageContent: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-628",
		},
	}

	for i, testCase := range testCases {
		email.ClearTestMessages()

		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.CourseMessageRecipients), &fields)
		fields["subject"] = testCase.Subject
		fields["body"] = testCase.Body
		fields["html"] = testCase.HTML
		fields["dry-run"] = testCase.DryRun

		response := core.SendTestAPIRequestFull(test, "courses/admin/email", fields, nil, testCase.UserEmail)
		if !response.Success {
			if testCase.Locator != response.Locator {
				test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
					i, testCase.Locator, response.Locator)
			}

			continue
		}

		if testCase.Locator != "" {
			test.Errorf("Case %d: Did not get an expected error. Expected: '%s'", i, testCase.Locator)
			continue
		}

		testMessages := email.GetTestMessages()

		if testCase.DryRun {
			if len(testMessages) != 0 {
				test.Errorf("Case %d: Unexpected emails sent. Expected '0 emails', found '%d emails'.", i, len(testMessages))
			}

			continue
		}

		if len(testMessages) != 1 {
			test.Errorf("Case %d: Unexpected number of emails sent. Expected '1 email', found '%d emails'.", i, len(testMessages))
			continue
		}

		var responseContent EmailResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if (len(responseContent.To) + len(responseContent.CC) + len(responseContent.BCC)) == 0 {
			test.Errorf("Case %d: Email was sent without any recipients.", i)
			continue
		}

		testMessage := testMessages[0]

		if testCase.Body != testMessage.Body {
			test.Errorf("Case %d: Unexpected body content. Expected: '%s', actual: '%s'.", i, testCase.Body, testMessage.Body)
			continue
		}

		if testCase.HTML != testMessage.HTML {
			test.Errorf("Case %d: Unexpected HTML value. Expected: '%t', actual: '%t'.", i, testCase.HTML, testMessage.HTML)
			continue
		}
	}
}
