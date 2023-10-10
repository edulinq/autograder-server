package email

import (
    "fmt"
    "net/smtp"
    "strings"

    "github.com/eriq-augustine/autograder/config"
)

func Send(to []string, subject string, body string, html bool) error {
    auth := smtp.PlainAuth("", config.EMAIL_USER.GetString(), config.EMAIL_PASS.GetString(), config.EMAIL_HOST.GetString());
    serverAddress := fmt.Sprintf("%s:%s", config.EMAIL_HOST.GetString(), config.EMAIL_PORT.GetString());

    contentType := "text/plain";
    if (html) {
        contentType = "text/html";
    }

    message := []byte(fmt.Sprintf(
        "To: %s\r\n" +
        "Subject: %s\r\n" +
        "Mime-Version: 1.0\r\n" +
        "Content-Transfer-Encoding: quoted-printable\r\n" +
        "Content-Type: %s; charset=UTF-8\r\n" +
        "\r\n" +
        "%s",
        strings.Join(to, ", "),
        subject,
        contentType,
        body));


    return smtp.SendMail(serverAddress, auth, config.EMAIL_FROM.GetString(), to, message);
}
