package email

import (
    "fmt"
    "net/smtp"
    "strings"

    "github.com/eriq-augustine/autograder/config"
)

func Send(to []string, subject string, body string) error {
    auth := smtp.PlainAuth("", config.EMAIL_USER.GetString(), config.EMAIL_PASS.GetString(), config.EMAIL_HOST.GetString());
    serverAddress := fmt.Sprintf("%s:%s", config.EMAIL_HOST.GetString(), config.EMAIL_PORT.GetString());

    message := []byte(fmt.Sprintf("" +
        "To: %s\r\n" +
        "Subject: %s\r\n" +
        "\r\n" +
        "%s\r\n",
        strings.Join(to, ", "),
        subject,
        body));


    return smtp.SendMail(serverAddress, auth, config.EMAIL_FROM.GetString(), to, message);
}
