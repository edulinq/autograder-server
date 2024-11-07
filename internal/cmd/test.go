package cmd

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/api/server"
	procedures "github.com/edulinq/autograder/internal/procedures/server"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/exit"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

const (
	TESTING_ARG0    = "testing"
	STDOUT_FILENAME = "stdout.txt"
	STDERR_FILENAME = "stderr.txt"
)

var (
	cmd_testing = false
	old_port    = 0
)

type CommonCMDTestCase struct {
	ExpectedExitCode        int
	ExpectedStdout          string
	ExpectedStderrSubstring string
	LogLevel                log.LogLevel
}

// Common setup for all CMD tests that require a server.
func CMDServerTestingMain(suite *testing.M) {
	server.StopServer()

	port, err := getUnusedPort()
	if err != nil {
		log.Fatal("Failed to get an unused port.", err)
	}

	// Run inside a func so defers will run before exit.Exit().
	code := func() int {
		defer config.WEB_PORT.Set(config.WEB_PORT.Get())
		config.WEB_PORT.Set(port)
		cmd_testing = true

		db.PrepForTestingMain()
		defer db.CleanupTestingMain()

		var serverRun sync.WaitGroup
		serverRun.Add(1)

		go func() {
			serverRun.Done()

			err := server.RunServer()
			if err != nil {
				log.Fatal("Failed to run the server.", err)
			}
		}()

		defer func() {
			server.StopServer()
			cmd_testing = false
		}()

		serverRun.Wait()

		return suite.Run()
	}()

	exit.Exit(code)
}

func getUnusedPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

func MustStartCMDServer() {
	// Don't start the server during tests (since tests already start their own server).
	if cmd_testing {
		return
	}

	port, err := getUnusedPort()
	if err != nil {
		log.Fatal("Failed to get an unused port.", err)
	}

	old_port = config.WEB_PORT.Get()
	config.WEB_PORT.Set(port)

	go func() {
		err = procedures.Start()
		if err != nil {
			log.Fatal("Failed to start the server.", err)
		}
	}()

	// Small sleep to allow the server to start up.
	time.Sleep(100 * time.Millisecond)
}

func MustStopCMDServer() {
	// Don't stop the server during tests (since tests already stop their server).
	if cmd_testing {
		return
	}

	server.StopServer()
	config.WEB_PORT.Set(old_port)
}

func RunCMDTest(test *testing.T, mainFunc func(), args []string, logLevel log.LogLevel) (string, string, int, error) {
	// Suppress exits to capture exit codes.
	exit.SetShouldExitForTesting(false)
	defer exit.SetShouldExitForTesting(true)

	tempDir := util.MustMkDirTemp("autograder-testing-cmd-")
	stdoutPath := filepath.Join(tempDir, STDOUT_FILENAME)
	stderrPath := filepath.Join(tempDir, STDERR_FILENAME)

	// Add in a dummy first arg.
	args = append([]string{TESTING_ARG0}, args...)

	// Ensure log levels are reset to their original state.
	oldTextLogLevel := log.GetTextLevel()
	oldBackendLogLevel := log.GetBackendLevel()
	defer func() {
		log.SetTextLevel(oldTextLogLevel)
		log.SetBackendLevel(oldBackendLogLevel)
	}()

	args = append(args, "--log-level", logLevel.String())

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
		log.SetTextWriter(oldStderr)
	}()

	// Capture the logging and stderr output to the same file.
	stderrFile := util.MustCreateFile(stderrPath)
	os.Stderr = stderrFile
	log.SetTextWriter(stderrFile)
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

	exitCode := exit.GetLastExitCode()

	return stdout, stderr, exitCode, err
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

func RunCommonCMDTests(test *testing.T, mainFunc func(), args []string, commonTestCase CommonCMDTestCase, prefix string) (string, string, int, bool) {
	stdout, stderr, exitCode, err := RunCMDTest(test, mainFunc, args, commonTestCase.LogLevel)
	if err != nil {
		test.Errorf("%sCMD run returned an error: '%v'.", prefix, err)
		return "", "", -1, false
	}

	if commonTestCase.ExpectedExitCode != exitCode {
		test.Errorf("%sUnexpected exit code. Expected: '%d', Actual: '%d'.", prefix, commonTestCase.ExpectedExitCode, exitCode)
		return "", "", -1, false
	}

	if !strings.Contains(stderr, commonTestCase.ExpectedStderrSubstring) {
		test.Errorf("%sUnexpected stderr substring. Expected stderr substring: '%s', Actual stderr: '%s'.", prefix, commonTestCase.ExpectedStderrSubstring, stderr)
		return "", "", -1, false
	}

	if commonTestCase.ExpectedStdout != stdout {
		test.Errorf("%sUnexpected output. Expected: \n'%s', \n Actual: \n'%s'.", prefix, commonTestCase.ExpectedStdout, stdout)
		return "", "", -1, false
	}

	return stdout, stderr, exitCode, true
}
