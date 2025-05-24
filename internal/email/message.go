package email

import (
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

type MessageRecipients struct {
	To  []string `json:"to"`
	CC  []string `json:"cc"`
	BCC []string `json:"bcc"`
}

type MessageContent struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	HTML    bool   `json:"html"`
}

type Message struct {
	MessageRecipients
	MessageContent
}

// Get the raw email formatted string (bytes).
func (this *Message) ToContent() []byte {
	contentType := "text/plain"
	if this.HTML {
		contentType = "text/html"
	}

	content := []byte(fmt.Sprintf(
		"To: %s\r\n"+
			"Cc: %s\r\n"+
			"Subject: %s\r\n"+
			"Mime-Version: 1.0\r\n"+
			"Content-Transfer-Encoding: quoted-printable\r\n"+
			"Content-Type: %s; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		strings.Join(this.To, ", "),
		strings.Join(this.CC, ", "),
		this.Subject,
		contentType,
		this.Body))

	return content
}

func (this *Message) LogValue() []*log.Attr {
	return []*log.Attr{
		log.NewAttr("to", this.To),
		log.NewAttr("cc", this.CC),
		log.NewAttr("bcc", this.BCC),
		log.NewAttr("subject", this.Subject),
	}
}

// An equality check (mainly for testing) that ignores the body.
func ShallowEqual(a *Message, b *Message) bool {
	if a == b {
		return true
	}

	if (a == nil) || (b == nil) {
		return false
	}

	if a.HTML != b.HTML {
		return false
	}

	if a.Subject != b.Subject {
		return false
	}

	return util.StringsEqualsIgnoreOrdering(a.To, b.To) &&
		util.StringsEqualsIgnoreOrdering(a.CC, b.CC) &&
		util.StringsEqualsIgnoreOrdering(a.BCC, b.BCC)
}

func Compare(a *Message, b *Message) int {
	if a == b {
		return 0
	}

	if a == nil {
		return 1
	} else if b == nil {
		return -1
	}

	value := strings.Compare(a.Subject, b.Subject)
	if value != 0 {
		return value
	}

	value = strings.Compare(a.Body, b.Body)
	if value != 0 {
		return value
	}

	value = util.StringsCompareIgnoreOrdering(a.To, b.To)
	if value != 0 {
		return value
	}

	value = util.StringsCompareIgnoreOrdering(a.CC, b.CC)
	if value != 0 {
		return value
	}

	value = util.StringsCompareIgnoreOrdering(a.BCC, b.BCC)
	if value != 0 {
		return value
	}

	return 0
}

func ShallowSliceEqual(a []*Message, b []*Message) bool {
	if len(a) != len(b) {
		return false
	}

	aSorted := a
	bSorted := b

	slices.SortFunc(aSorted, Compare)
	slices.SortFunc(bSorted, Compare)

	return slices.EqualFunc(aSorted, bSorted, ShallowEqual)
}

func (this *MessageRecipients) IsEmpty() bool {
	return (len(this.To) + len(this.CC) + len(this.BCC)) == 0
}
