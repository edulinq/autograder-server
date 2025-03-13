package email

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestSendEmailFullBase(test *testing.T) {
	ClearTestMessages()
	defer ClearTestMessages()

	message := &Message{
		To:      []string{"a@test.edulinq.org"},
		CC:      []string{"b@test.edulinq.org"},
		BCC:     []string{"c@test.edulinq.org"},
		Subject: "sub",
		Body:    "body",
		HTML:    false,
	}

	err := SendFull(message.To, message.CC, message.BCC, message.Subject, message.Body, message.HTML)
	if err != nil {
		test.Fatalf("Failed to send email: '%v'.", err)
	}

	expected := []*Message{message}
	actual := GetTestMessages()

	if !reflect.DeepEqual(expected, GetTestMessages()) {
		test.Fatalf("Messages not as expected. Expected: '%s', Actual: '%s'.", util.MustToJSONIndent(expected), util.MustToJSONIndent(actual))
	}
}
