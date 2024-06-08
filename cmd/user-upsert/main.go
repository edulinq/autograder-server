package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/procedures"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs

	Email       string `help:"Email for the user." arg:"" required:""`
	Name        string `help:"Name for the user."`
	Role        string `help:"Server role for the user. Defaults to 'user'." default:"user"`
	Pass        string `help:"Password for the user. Defaults to a random string (will be output)."`
	Course      string `help:"Optional ID of course to enroll user in."`
	CourseRole  string `help:"Role for the new user in the specified course. Defaults to 'student'." default:"student"`
	CourseLMSID string `help:"LMS ID for the new user in the specified course."`

	Force     bool `help:"Overwrite any existing user." default:"false"`
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

	data := procedures.RawUserUpsertData{
		Email:       args.Email,
		Name:        args.Name,
		Role:        args.Role,
		Pass:        args.Pass,
		Course:      args.Course,
		CourseRole:  args.CourseRole,
		CourseLMSID: args.CourseLMSID,
	}

	options := procedures.UpsertUsersOptions{
		RawUsers:          []AddUserData{data},
		Force:             args.Force,
		SendEmails:        args.SendEmail,
		DryRun:            args.DryRun,
		ContextServerRole: ServerRoleRoot,
	}

	result := procedures.UpsertUser(options)

	fmt.Println(util.MustToJSONIndent(result))

	if result.HasErrors() {
		os.Exit(1)
	}
}
