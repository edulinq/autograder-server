package main

import (
	"os"

	"github.com/alecthomas/kong"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	pcourses "github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

var args struct {
	config.ConfigArgs
}

func main() {
	kong.Parse(&args)
	err := config.HandleConfigArgs(args.ConfigArgs)
	if err != nil {
		log.Fatal("Could not load config options.", err)
	}

	log.Info("Autograder Version", log.NewAttr("version", util.GetAutograderFullVersion()))

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get working directory.", err)
	}

	db.MustOpen()
	defer db.MustClose()

	log.Info("Running server with working directory.", log.NewAttr("dir", workingDir))

	_, err = db.AddCourses()
	if err != nil {
		log.Fatal("Could not load courses.", err)
	}

	courses := db.MustGetCourses()
	log.Info("Loaded course(s).", log.NewAttr("count", len(courses)))

	// Startup courses (in the background).
	for _, course := range courses {
		log.Info("Loaded course.", course)
		go func(course *model.Course) {
			pcourses.UpdateCourse(course, true)
		}(course)
	}

	// Cleanup any temp dirs.
	defer util.RemoveRecordedTempDirs()

	err = api.StartServer()
	if err != nil {
		log.Fatal("Server was stopped.", err)
	}

	log.Info("Server closed.")
}
