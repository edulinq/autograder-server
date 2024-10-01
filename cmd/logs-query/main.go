package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/logs"
)

var args struct {
	config.ConfigArgs

	Level string `help:"Only includes logs from this level or higher." short:"l" default:"info"`
	Time  string `help:"Only includes logs from this time or later." short:"t"`

	Course     string `help:"Only includes logs from this course."`
	Assignment string `help:"Only includes logs from this assignment." short:"a"`
	User       string `help:"Only includes logs from this user." short:"u"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Dump all the loaded config and exit."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	user := db.MustGetServerUser(model.RootUserEmail)

	query := log.RawLogQuery{
		LevelString:  args.Level,
		AfterString:  args.Time,
		CourseID:     args.Course,
		AssignmentID: args.Assignment,
		TargetUser:   args.User,
	}

	records, locatableErr, err := logs.Query(query, user)
	if err != nil {
		log.Fatal("Failed to query logs.", err)
	}

	if locatableErr != nil {
		log.Fatal("Invalid logs query.", locatableErr)
	}

	for _, record := range records {
		fmt.Println(record.String())
	}
}
