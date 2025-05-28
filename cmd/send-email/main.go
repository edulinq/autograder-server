package main

import (
	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
)

var args struct {
	config.ConfigArgs
	To      []string `help:"Email (to)." required:""`
	CC      []string `help:"Email (cc)." optional:""`
	BCC     []string `help:"Email (bcc)." optional:""`
	Subject string   `help:"Email subject." required:""`
	Body    string   `help:"Email body." required:""`
}

func main() {
	kong.Parse(&args,
		kong.Description("Send an email."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	err = email.SendFull(args.To, args.CC, args.BCC, args.Subject, args.Body, false)
	if err != nil {
		log.Fatal("Could not send email.", err)
	}

	err = email.Close()
	if err != nil {
		log.Fatal("Failed to close SMTP connection.", err)
	}
}
