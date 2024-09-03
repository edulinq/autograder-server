package server

import (
	"fmt"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	pcourses "github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

func Start() error {
	defer api.StopServers()

	log.Info("Autograder Version.", log.NewAttr("version", util.GetAutograderFullVersion()))

	err := db.Open()
	if err != nil {
		log.Fatal("Failed to open the database.", err)
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Error("Failed to close the database.", err)
		}
	}()

	log.Info("Running server with working directory.", log.NewAttr("dir", config.GetWorkDir()))

	_, err = db.AddCourses()
	if err != nil {
		log.Fatal("Could not load courses.", err)
	}

	courses, err := db.GetCourses()
	if err != nil {
		log.Fatal("Failed to get courses.", err)
	}

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
		return fmt.Errorf("Failed to start server.")
	}

	log.Info("Server closed.")
	return nil
}
