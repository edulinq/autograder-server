package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/users"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs

	model.RawServerUserData

	SendEmail bool `help:"Send an email to the user if important changes (like a new password) were made." default:"false"`
	DryRun    bool `help:"Do not actually write out the user's file or send emails, just state what would happen." default:"false"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Upsert (update or insert) a user."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	// If a password was set on the command-line, then hash it before upserting.
	if args.RawServerUserData.Pass != "" {
		args.RawServerUserData.Pass = util.Sha256HexFromString(args.RawServerUserData.Pass)
	}

	options := users.UpsertUsersOptions{
		RawUsers:          []*model.RawServerUserData{&args.RawServerUserData},
		SendEmails:        args.SendEmail,
		DryRun:            args.DryRun,
		ContextServerRole: model.ServerRoleRoot,
	}

	result := users.UpsertUser(options)

	fmt.Println(util.MustToJSONIndent(result))

	if result.HasErrors() {
		os.Exit(1)
	}
}
