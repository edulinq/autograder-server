package server

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	pcourses "github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

func Start(initiator common.ServerInitiator) (err error) {
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

	defer util.RemoveRecordedTempDirs()

	err = server.RunServer(initiator)
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

// Check to see if a server is running and start one if it's not.
// Returns (false, 0) if a server is already running
// or (true, oldPort) if it started it's own server.
func MustEnsureServerIsRunning() (bool, int) {
	statusInfo, err := common.CheckAndHandleServerStatusFile()
	if err != nil {
		log.Fatal("Failed to retrieve the current status file's json.", err)
	}

	if statusInfo != nil {
		// Don't start the server if the primary server or cmd test server is running.
		if statusInfo.ServerInitiator == common.PRIMARY_SERVER || statusInfo.ServerInitiator == common.CMD_TEST_SERVER {
			return false, 0
		}
	}

	port, err := GetUnusedPort()
	if err != nil {
		log.Fatal("Failed to get an unused port.", err)
	}

	oldPort := config.WEB_PORT.Get()
	config.WEB_PORT.Set(port)

	var serverStart sync.WaitGroup
	serverStart.Add(1)

	go func() {
		serverStart.Done()

		err = Start(common.CMD_SERVER)
		if err != nil {
			log.Fatal("Failed to start the server.", err)
		}
	}()

	serverStart.Wait()

	// Small sleep to allow the server to start up.
	time.Sleep(100 * time.Millisecond)

	return true, oldPort
}

func GetUnusedPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}
