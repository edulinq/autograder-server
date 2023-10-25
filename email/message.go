package email

import (
    "fmt"
    "strings"
    "slices"
)

type Message struct {
    To []string `json:"to"`
    Subject string `json:"subject"`
    Body string `json:"body"`
    HTML bool `json:"html"`
}

// Get the raw email formatted string (bytes).
func (this *Message) ToContent() []byte {
    contentType := "text/plain";
    if (this.HTML) {
        contentType = "text/html";
    }

    content := []byte(fmt.Sprintf(
        "To: %s\r\n" +
        "Subject: %s\r\n" +
        "Mime-Version: 1.0\r\n" +
        "Content-Transfer-Encoding: quoted-printable\r\n" +
        "Content-Type: %s; charset=UTF-8\r\n" +
        "\r\n" +
        "%s",
        strings.Join(this.To, ", "),
        this.Subject,
        contentType,
        this.Body));

    return content;
}

// An equality check (mainly for testing) that ignores the body.
func ShallowEqual(a *Message, b *Message) bool {
    if (a == b) {
        return true;
    }

    if ((a == nil) || (b == nil)) {
        return false;
    }

    aTo := a.To[:];
    bTo := b.To[:];

    slices.Sort(aTo);
    slices.Sort(bTo);

    return slices.Equal(aTo, bTo) && (a.Subject == b.Subject) && (a.HTML == b.HTML);
}
