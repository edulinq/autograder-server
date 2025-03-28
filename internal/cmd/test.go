package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/exit"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/procedures/server"
	"github.com/edulinq/autograder/internal/systemserver"
	"github.com/edulinq/autograder/internal/util"
)

const (
	TESTING_ARG0    = "testing"
	STDOUT_FILENAME = "stdout.txt"
	STDERR_FILENAME = "stderr.txt"
)

type CommonCMDTestCase struct {
	ExpectedExitCode        int
	ExpectedStdout          string
	IgnoreStdout            bool
	ExpectedStderrSubstring string
	LogLevel                log.LogLevel
}

// Common setup for all CMD tests that require a server.
func CMDServerTestingMain(suite *testing.M) {
	err := server.CleanupAndStop()
	if err != nil {
		log.Fatal("Failed to cleanup and stop server before running the CMD test server.", err)
	}

	port, err := util.GetUnusedPort()
	if err != nil {
		log.Fatal("Failed to get an unused port.", err)
	}

	// Run inside a func so defers will run before exit.Exit().
	code := func() int {
		defer config.WEB_HTTP_PORT.Set(config.WEB_HTTP_PORT.Get())
		config.WEB_HTTP_PORT.Set(port)

		db.PrepForTestingMain()
		defer db.CleanupTestingMain()

		var serverRun sync.WaitGroup
		serverRun.Add(1)

		go func() {
			serverRun.Done()

			err := server.RunAndBlockFull(systemserver.CMD_TEST_SERVER, true)
			if err != nil {
				log.Fatal("Failed to run the server.", err)
			}
		}()

		defer func() {
			err := server.CleanupAndStop()
			if err != nil {
				log.Fatal("Failed to cleanup and stop the CMD test server.", err)
			}
		}()

		serverRun.Wait()

		// Small sleep to allow the server to start up.
		time.Sleep(150 * time.Millisecond)

		return suite.Run()
	}()

	exit.Exit(code)
}

func RunCMDTest(test *testing.T, mainFunc func(), args []string, logLevel log.LogLevel) (string, string, int, error) {
	// Suppress exits to capture exit codes.
	exit.SetShouldExitForTesting(false)
	defer exit.SetShouldExitForTesting(true)

	tempDir := util.MustMkDirTemp("autograder-testing-cmd-")
	defer util.RemoveDirent(tempDir)

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

	os.Stdout = mustCreateFile(stdoutPath)

	// Setup stderr capture.
	oldStderr := os.Stderr
	defer func() {
		os.Stderr = oldStderr
		log.SetTextWriter(oldStderr)
	}()

	// Capture the logging and stderr output to the same file.
	stderrFile := mustCreateFile(stderrPath)
	os.Stderr = stderrFile
	log.SetTextWriter(stderrFile)

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
		logOutputs(test, stdout, stderr)
		return "", "", -1, false
	}

	if commonTestCase.ExpectedExitCode != exitCode {
		test.Errorf("%sUnexpected exit code. Expected: '%d', Actual: '%d'.", prefix, commonTestCase.ExpectedExitCode, exitCode)
		logOutputs(test, stdout, stderr)
		return "", "", -1, false
	}

	if !strings.Contains(stderr, commonTestCase.ExpectedStderrSubstring) {
		test.Errorf("%sUnexpected stderr substring. Expected stderr substring: '%s', Actual stderr: '%s'.", prefix, commonTestCase.ExpectedStderrSubstring, stderr)
		logOutputs(test, stdout, stderr)
		return "", "", -1, false
	}

	if !commonTestCase.IgnoreStdout && commonTestCase.ExpectedStdout != stdout {
		test.Errorf("%sUnexpected output. Expected: \n'%s', \n Actual: \n'%s'.", prefix, commonTestCase.ExpectedStdout, stdout)
		logOutputs(test, stdout, stderr)
		return "", "", -1, false
	}

	return stdout, stderr, exitCode, true
}

func logOutputs(test *testing.T, stdout string, stderr string) {
	test.Log("--- stdout ---")
	test.Log(stdout)
	test.Log("--------------")
	test.Log("--- stderr ---")
	test.Log(stderr)
	test.Log("--------------")
}

func mustCreateFile(path string) *os.File {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal("Unable to create file.", err, log.NewAttr("path", path))
	}

	return file
}
