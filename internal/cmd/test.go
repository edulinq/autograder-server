package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/api/server"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const (
	TESTING_ARG0    = "testing"
	STDOUT_FILENAME = "stdout.txt"
	STDERR_FILENAME = "stderr.txt"
)

// Common setup for all CMD tests that require a server.
func CMDServerTestingMain(suite *testing.M) {
	// Run inside a func so defers will run before os.Exit().
	code := func() int {
		db.PrepForTestingMain()
		defer db.CleanupTestingMain()

		log.SetLevelDebug()

		var serverRun sync.WaitGroup
		serverRun.Add(1)

		config.NO_AUTH.Set(false)

		go func() {
			serverRun.Done()
			server.StartServer()
		}()
		defer server.StopServer()

		serverRun.Wait()

		// Small sleep to allow the server to start up.
		time.Sleep(100 * time.Millisecond)

		return suite.Run()
	}()

	os.Exit(code)
}

func RunCMDTest(test *testing.T, mainFunc func(), args []string) (string, string, error) {
	tempDir := util.MustMkDirTemp("autograder-testing-cmd-")
	stdoutPath := filepath.Join(tempDir, STDOUT_FILENAME)
	stderrPath := filepath.Join(tempDir, STDERR_FILENAME)

	// Add in a dummy first arg.
	args = append([]string{TESTING_ARG0}, args...)

	// Setup stdout capture.
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
	}()

	os.Stdout = util.MustCreateFile(stdoutPath)

	// Setup stderr capture.
	oldStderr := os.Stderr
	defer func() {
		os.Stderr = oldStderr
	}()

	os.Stderr = util.MustCreateFile(stderrPath)

	// Run.
	err := runCMD(mainFunc, args)

	// Put back stdout.
	os.Stdout = oldStdout
	stdout := util.MustReadFile(stdoutPath)

	// Put back stderr.
	os.Stderr.Close()
	os.Stderr = oldStderr
	stderr := util.MustReadFile(stderrPath)

	return stdout, stderr, err
}

func runCMD(mainFunc func(), args []string) (err error) {
	err = nil

	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = fmt.Errorf("Main function panicked: '%v'.", value)
	}()

	// Set new args.
	oldArgs := os.Args
	os.Args = append([]string(nil), args...)

	// Reset the existing args.
	defer func() {
		os.Args = oldArgs
	}()

	// Capture fatal logs.
	oldPanicValue := log.SetPanicOnFatal(true)
	defer log.SetPanicOnFatal(oldPanicValue)

	mainFunc()

	return err
}
