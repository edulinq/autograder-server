package server

import (
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	pcourses "github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

type ProcedureServer struct {
	apiServer *server.APIServer
}

var serverInstance *ProcedureServer

func setupServer(initiator common.ServerInitiator) error {
	version, err := util.GetAutograderVersion()
	if err != nil {
		log.Warn("Failed to get the autograder version.", err)
	}

	log.Info("Autograder Version.", log.NewAttr("version", version))

	err = db.Open()
	if err != nil {
		return fmt.Errorf("Failed to open the database: '%w'.", err)
	}

	log.Debug("Setup server with working directory.", log.NewAttr("dir", config.GetWorkDir()))

	courses, err := db.GetCourses()
	if err != nil {
		return fmt.Errorf("Failed to get courses: '%w'.", err)
	}

	log.Debug("Found course(s).", log.NewAttr("count", len(courses)))

	// Startup courses (in the background).
	for _, course := range courses {
		go startCourse(course)
	}

	return nil
}

func CleanupAndStopServer() (err error) {
	if serverInstance == nil {
		return nil
	}

	err = errors.Join(err, db.Close())
	err = errors.Join(err, util.RemoveRecordedTempDirs())

	serverInstance.apiServer.StopServer()

	serverInstance.apiServer = nil
	serverInstance = nil

	return err
}

func RunAndBlockServer(initiator common.ServerInitiator, skipSetup bool) (err error) {
	// Run inside a func so defers will run before the function returns.
	func() {
		defer func() {
			err = errors.Join(err, CleanupAndStopServer())
		}()

		serverInstance = &ProcedureServer{
			apiServer: server.NewAPIServer(),
		}

		if !skipSetup {
			err = setupServer(initiator)
			if err != nil {
				err = fmt.Errorf("Failed to setup the server: '%w'.", err)
				return
			}
		}

		err = serverInstance.apiServer.RunServer(initiator)
		if err != nil {
			err = fmt.Errorf("API server run returned an error: '%w'.", err)
		}
	}()

	return err
}

func startCourse(course *model.Course) {
	root, err := db.GetRoot()
	if err != nil {
		log.Error("Failed to get root for course update.", err, course)
		return
	}

	options := pcourses.CourseUpsertOptions{
		ContextUser: root,
	}

	_, err = pcourses.UpdateFromLocalSource(course, options)
	if err != nil {
		log.Error("Failed to update course.", err, course)
	}
}
