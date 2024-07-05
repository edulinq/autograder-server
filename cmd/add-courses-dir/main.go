package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
)

var args struct {
	config.ConfigArgs
	Path string `help:"Path to course JSON file." arg:"" type:"existingdir"`
}

func main() {
	kong.Parse(&args,
		kong.Description("Add all specified courses to the system."),
	)

	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	courseIDs, err := db.AddCoursesFromDir(args.Path, nil)
	if err != nil {
		log.Fatal("Could not add courses dir.", log.NewAttr("path", args.Path), err)
	}

	fmt.Printf("Added %d courses.\n", len(courseIDs))
	for _, courseID := range courseIDs {
		fmt.Printf("    %s\n", courseID)
	}
}
