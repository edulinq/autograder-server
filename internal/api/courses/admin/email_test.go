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
		UserEmail  string
		Recipients model.CourseMessageRecipients
		Content    email.MessageContent
		Locator    string
		DryRun     bool
	}{
		// Valid Permissions
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
		},
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"student"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
		},
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"outside-email@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
		},

		// Valid Permissions, Role Escalation
		{
			UserEmail: "server-admin",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
		},
		{
			UserEmail: "server-owner",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
		},

		// CC
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				CC: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
		},

		// BCC
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				BCC: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
		},

		// Body
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
				Body:    "Body",
			},
		},

		// HTML
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
				HTML:    true,
			},
		},

		// Dry Run
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
			DryRun: true,
		},

		// Invalid Permissions
		{
			UserEmail: "course-student",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-020",
		},

		// Invalid Permissions, Role Escalation
		{
			UserEmail: "server-user",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-040",
		},
		{
			UserEmail: "server-creator",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-040",
		},

		// No Subject
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org"},
			},
			Locator: "-627",
		},

		// User Errors
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"creator"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-628",
		},

		// No Recipients
		{
			UserEmail: "course-grader",
			Content: email.MessageContent{
				Subject: "Message in a Bottle",
			},
			Locator: "-629",
		},

		// No recipients after resolving email addresses.
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"course-student@test.edulinq.org", "-course-student@test.edulinq.org"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-629",
		},
		{
			UserEmail: "course-grader",
			Recipients: model.CourseMessageRecipients{
				To: []model.CourseUserReference{"student", "-student"},
			},
			Content: email.MessageContent{
				Subject: "Subject",
			},
			Locator: "-629",
		},
	}

	for i, testCase := range testCases {
		email.ClearTestMessages()

		var fields map[string]any
		util.MustJSONFromString(util.MustToJSON(testCase.Recipients), &fields)
		fields["subject"] = testCase.Content.Subject
		fields["body"] = testCase.Content.Body
		fields["html"] = testCase.Content.HTML
		fields["dry-run"] = testCase.DryRun

		response := core.SendTestAPIRequestFull(test, "courses/admin/email", fields, nil, testCase.UserEmail)
		if !response.Success {
			if testCase.Locator != response.Locator {
				test.Errorf("Case %d: Incorrect error returned. Expected: '%s', Actual: '%s'.",
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
				test.Errorf("Case %d: Unexpected emails sent. Expected: '0 emails', Actual: '%d emails'.", i, len(testMessages))
			}

			continue
		}

		if len(testMessages) != 1 {
			test.Errorf("Case %d: Unexpected number of emails sent. Expected: '1 email', Actual: '%d emails'.", i, len(testMessages))
			continue
		}

		var responseContent EmailResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if (len(responseContent.To) + len(responseContent.CC) + len(responseContent.BCC)) == 0 {
			test.Errorf("Case %d: Email was sent without any recipients.", i)
			continue
		}

		testMessage := testMessages[0]

		if testCase.Content.Body != testMessage.Body {
			test.Errorf("Case %d: Unexpected body content. Expected: '%s', Actual: '%s'.", i, testCase.Content.Body, testMessage.Body)
			continue
		}

		if testCase.Content.HTML != testMessage.HTML {
			test.Errorf("Case %d: Unexpected HTML value. Expected: '%t', Actual: '%t'.", i, testCase.Content.HTML, testMessage.HTML)
			continue
		}
	}
}
