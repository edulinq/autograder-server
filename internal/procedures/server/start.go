package server

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	pcourses "github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

func Start() error {
	log.Info("Autograder Version", log.NewAttr("version", util.GetAutograderFullVersion()))

	var pidFilePath = config.PID_PATH.Get()

	if _, err := os.Stat(pidFilePath); err == nil {
		data, err := os.ReadFile(pidFilePath)
		if err == nil {
			pid, err := strconv.Atoi(string(data))
			if err == nil {
				process, err := os.FindProcess(pid)
				if err == nil {
					err = process.Signal(syscall.Signal(0))
					if err == nil {
						log.Fatal("Server is already running with PID:", pid)
					} else {
						os.Remove(pidFilePath)
					}
				}
			}
		}
	}

	err := common.CreatePIDFile()
	if err != nil {
		log.Fatal("Could not create PID file", err)
	}

	defer os.Remove(pidFilePath)

	// Remove the unix socket file when the program terminates abruptly.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(pidFilePath)
		os.Exit(1)
	}()

	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get working directory", err)
	}

	db.MustOpen()
	defer db.MustClose()

	log.Info("Running server with working directory.", log.NewAttr("dir", workingDir))

	_, err = db.AddCourses()
	if err != nil {
		log.Fatal("Could not load courses", err)
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
		log.Fatal("Server was stopped", err)
	}

	log.Info("Server closed.")
	return nil
}
