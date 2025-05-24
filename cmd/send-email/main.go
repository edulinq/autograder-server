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
	Course  string                      `help:"Optional Course ID. Only required when roles or * (all course users) are in the recipients." arg:"" optional:""`
	To      []model.CourseUserReference `help:"Email recipients (to)." required:""`
	CC      []model.CourseUserReference `help:"Email recipients (cc)." optional:""`
	BCC     []model.CourseUserReference `help:"Email recipients (bcc)." optional:""`
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

	db.MustOpen()
	defer db.MustClose()

	var to []string = nil
	var cc []string = nil
	var bcc []string = nil

	if args.Course != "" {
		course := db.MustGetCourse(args.Course)

		users, err := db.GetCourseUsers(course)
		if err != nil {
			log.Fatal("Failed to get course users.", err)
		}

		courseRecipients := model.CourseMessageRecipients{
			To:  args.To,
			CC:  args.CC,
			BCC: args.BCC,
		}

		recipients, userErrors := courseRecipients.ToMessageRecipients(users)
		if err != nil {
			log.Fatal("Failed to resolve users.", log.NewAttr("errors", userErrors))
		}

		to = recipients.To
		cc = recipients.CC
		bcc = recipients.BCC
	} else {
		to = make([]string, 0, len(args.To))
		for _, email := range args.To {
			to = append(to, string(email))
		}

		cc = make([]string, 0, len(args.CC))
		for _, email := range args.CC {
			cc = append(cc, string(email))
		}

		bcc = make([]string, 0, len(args.BCC))
		for _, email := range args.BCC {
			bcc = append(bcc, string(email))
		}
	}

	err = email.SendFull(to, cc, bcc, args.Subject, args.Body, false)
	if err != nil {
		log.Fatal("Could not send email.", err)
	}

	err = email.Close()
	if err != nil {
		log.Fatal("Failed to close SMTP connection.", err)
	}
}
