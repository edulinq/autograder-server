package email

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/timestamp"
)

const MIN_TIME_BETWEEN_EMAILS_MSEC = 750
const LOCK_KEY = "internal.email"

var lastEmailTime timestamp.Timestamp = timestamp.Zero()

// Messages that are stored (instead of sent) in testing mode.
var testMessages []*Message = nil

func Send(to []string, subject string, body string, html bool) error {
	return SendMessage(&Message{
		To:      to,
		Subject: subject,
		Body:    body,
		HTML:    html,
	})
}

func SendMessage(message *Message) error {
	auth := smtp.PlainAuth("", config.EMAIL_USER.Get(), config.EMAIL_PASS.Get(), config.EMAIL_HOST.Get())

	serverAddress := fmt.Sprintf("%s:%s", config.EMAIL_HOST.Get(), config.EMAIL_PORT.Get())
	content := message.ToContent()

	// In testing mode, just store the message.
	if config.TESTING_MODE.Get() {
		testMessages = append(testMessages, message)
		return nil
	}

	// Only send one email at a time.
	common.Lock(LOCK_KEY)
	defer common.Unlock(LOCK_KEY)

	// Sleep if we are sending emails too fast.
	timeSinceLastEmail := timestamp.Now() - lastEmailTime
	sleepDuration := timestamp.FromMSecs(MIN_TIME_BETWEEN_EMAILS_MSEC) - timeSinceLastEmail
	if sleepDuration > 0 {
		time.Sleep(sleepDuration.ToGoTimeDuration())
	}

	lastEmailTime = timestamp.Now()

	return smtp.SendMail(serverAddress, auth, config.EMAIL_FROM.Get(), message.To, content)
}

func GetTestMessages() []*Message {
	return testMessages
}

func ClearTestMessages() {
	testMessages = nil
}
