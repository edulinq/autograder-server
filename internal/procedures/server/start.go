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

func Start(creator string) (err error) {
	defer server.StopServer()

	version, err := util.GetAutograderVersion()
	if err != nil {
		log.Warn("Failed to get the autograder version.", err)
	}

	log.Info("Autograder Version.", log.NewAttr("version", version))

	err = db.Open()
	if err != nil {
		return fmt.Errorf("Failed to open the database: '%w'.", err)
	}

	defer func() {
		err = errors.Join(err, db.Close())
	}()

	log.Debug("Running server with working directory.", log.NewAttr("dir", config.GetWorkDir()))

	courses, err := db.GetCourses()
	if err != nil {
		return fmt.Errorf("Failed to get courses: '%w'.", err)
	}

	log.Debug("Found course(s).", log.NewAttr("count", len(courses)))

	// Startup courses (in the background).
	for _, course := range courses {
		go startCourse(course)
	}

	// Cleanup any temp dirs.
	defer util.RemoveRecordedTempDirs()

	err = server.RunServer(creator)
	if err != nil {
		return fmt.Errorf("Error during server startup sequence: '%w'.", err)
	}

	log.Debug("Server closed.")

	return err
}

func startCourse(course *model.Course) {
	root, err := db.GetRoot()
	if err != nil {
		log.Error("Failed to get root for course update.", err, course)
	}

	options := pcourses.CourseUpsertOptions{
		ContextUser: root,
	}

	_, err = pcourses.UpdateFromLocalSource(course, options)
	if err != nil {
		log.Error("Failed to update course.", err, course)
	}
}
