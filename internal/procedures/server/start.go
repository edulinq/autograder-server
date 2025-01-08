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
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/task"
	"github.com/edulinq/autograder/internal/util"
)

const SERVER_LOCK = "internal.procedures.server.SERVER_LOCK"

var apiServer *server.APIServer = nil

func setup(initiator common.ServerInitiator) error {
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

	// Start stat collection.
	stats.StartCollection(config.STATS_SYSTEM_INTERVAL_MS.Get())

	courses, err := db.GetCourses()
	if err != nil {
		return fmt.Errorf("Failed to get courses: '%w'.", err)
	}

	log.Debug("Found course(s).", log.NewAttr("count", len(courses)))

	// Only perfrom some tasks if we are running a primary server.
	if initiator == common.PRIMARY_SERVER {
		// Initialize the task engine.
		task.Init()

		// Startup courses (in the background).
		for _, course := range courses {
			go startCourse(course)
		}
	}

	return nil
}

func CleanupAndStop() (err error) {
	common.Lock(SERVER_LOCK)
	defer common.Unlock(SERVER_LOCK)

	if apiServer == nil {
		return nil
	}

	stats.StopCollection()

	apiServer.Stop()
	apiServer = nil

	err = errors.Join(err, db.Close())
	err = errors.Join(err, util.RemoveRecordedTempDirs())

	log.Debug("Server closed.")

	return err
}

func assignAndSetupServer(initiator common.ServerInitiator, skipSetup bool) error {
	common.Lock(SERVER_LOCK)
	defer common.Unlock(SERVER_LOCK)

	apiServer = server.NewAPIServer()

	if !skipSetup {
		err := setup(initiator)
		if err != nil {
			return fmt.Errorf("Failed to setup the server: '%w'.", err)
		}
	}

	return nil
}

func RunAndBlock(initiator common.ServerInitiator) (err error) {
	return RunAndBlockFull(initiator, false)
}

func RunAndBlockFull(initiator common.ServerInitiator, skipSetup bool) (err error) {
	// Run inside a func so defers will run before the function returns.
	func() {
		defer func() {
			err = errors.Join(err, CleanupAndStop())
		}()

		err = assignAndSetupServer(initiator, skipSetup)
		if err != nil {
			err = fmt.Errorf("Failed to assign and setup server: '%w'.", err)
			return
		}

		// apiServer may be nil after this call completes if CleanupAndStop() is called concurrently.
		err = apiServer.RunAndBlock(initiator)
		if err != nil {
			err = fmt.Errorf("API server run returned an error: '%w'.", err)
			return
		}
	}()

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
