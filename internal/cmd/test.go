package cmd

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/api"
	"github.com/edulinq/autograder/internal/api/core"
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

var server *httptest.Server
var oldServerPort int

// Common setup for all CMD tests that require a server.
func CMDServerTestingMain(suite *testing.M) {
	// Run inside a func so defers will run before os.Exit().
	code := func() int {
		db.PrepForTestingMain()
		defer db.CleanupTestingMain()

		config.NO_AUTH.Set(false)

		startTestServer()
		defer stopTestServer()

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

	// Run
	err := runCMD(mainFunc, args)

	// Put back stdout.
	os.Stdout.Close()
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

func startTestServer() {
	if server != nil {
		panic("Test server already started.")
	}

	server = httptest.NewServer(core.GetRouteServer(api.GetRoutes()))

	parts := strings.Split(server.URL, ":")
	newPort, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		panic(fmt.Sprintf("Cannot get port from server URL ('%s'): '%v'.", server.URL, err))
	}

	oldServerPort = config.WEB_PORT.Get()
	config.WEB_PORT.Set(newPort)
}

func stopTestServer() {
	if server != nil {
		server.Close()

		server = nil
		config.WEB_PORT.Set(oldServerPort)
	}
}
