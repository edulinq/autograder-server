package server

import (
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	pcourses "github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

func Start() (err error) {
	defer server.StopServer()

	log.Info("Autograder Version.", log.NewAttr("version", util.GetAutograderFullVersion()))

	err = db.Open()
	if err != nil {
		return fmt.Errorf("Failed to open the database: '%w'.", err)
	}

	defer func() {
		err = errors.Join(err, db.Close())
	}()

	log.Debug("Running server with working directory.", log.NewAttr("dir", config.GetWorkDir()))

	_, err = db.AddCourses()
	if err != nil {
		return fmt.Errorf("Failed to load courses: '%w'.", err)
	}

	courses, err := db.GetCourses()
	if err != nil {
		return fmt.Errorf("Failed to get courses: '%w'.", err)
	}

	log.Debug("Loaded course(s).", log.NewAttr("count", len(courses)))

	// Startup courses (in the background).
	for _, course := range courses {
		go func(course *model.Course) {
			pcourses.UpdateCourse(course, true)
		}(course)
	}

	// Cleanup any temp dirs.
	defer util.RemoveRecordedTempDirs()

	err = server.RunServer()
	if err != nil {
		return fmt.Errorf("Error during server startup sequence: '%w'.", err)
	}

	log.Debug("Server closed.")

	return err
}
