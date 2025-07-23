package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

var args struct {
	config.ConfigArgs
	To      []model.ServerUserReference `help:"Email (to). Accepts server user references." required:""`
	CC      []model.ServerUserReference `help:"Email (cc). Accepts server user references." optional:""`
	BCC     []model.ServerUserReference `help:"Email (bcc). Accepts server user references." optional:""`
	Subject string                      `help:"Email subject." required:""`
	Body    string                      `help:"Email body." required:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Send an email."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	courses := db.MustGetCourses()
	users := db.MustGetServerUsers()

	serverRecipients := model.ServerMessageRecipients{
		To:  args.To,
		CC:  args.CC,
		BCC: args.BCC,
	}

	recipients, err := serverRecipients.ToMessageRecipients(courses, users)
	if err != nil {
		log.Fatal("Failed to convert references to message recipients.", err)
	}

	err = email.SendFull(recipients.To, recipients.CC, recipients.BCC, args.Subject, args.Body, false)
	if err != nil {
		log.Fatal("Could not send email.", err)
	}

	err = email.Close()
	if err != nil {
		log.Fatal("Failed to close SMTP connection.", err)
	}
}
