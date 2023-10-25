package email

import (
    "fmt"
    "net/smtp"

    "github.com/eriq-augustine/autograder/config"
)

// Messages that are stored (instead of sent) in testing mode.
var testMessages []*Message = nil;

func Send(to []string, subject string, body string, html bool) error {
    return SendMessage(&Message{
        To: to,
        Subject: subject,
        Body: body,
        HTML: html,
    });
}

func SendMessage(message *Message) error {
    auth := smtp.PlainAuth("", config.EMAIL_USER.GetString(), config.EMAIL_PASS.GetString(), config.EMAIL_HOST.GetString());

    serverAddress := fmt.Sprintf("%s:%s", config.EMAIL_HOST.GetString(), config.EMAIL_PORT.GetString());
    content := message.ToContent();

    // In testing mode, just store the message.
    if (config.TESTING_MODE.GetBool()) {
        testMessages = append(testMessages, message);
        return nil;
    }

    return smtp.SendMail(serverAddress, auth, config.EMAIL_FROM.GetString(), message.To, content);
}

func GetTestMessages() []*Message {
    return testMessages;
}

func ClearTestMessages() {
    testMessages = nil;
}
