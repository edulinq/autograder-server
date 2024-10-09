package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/api/server"
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

		var serverRun sync.WaitGroup
		serverRun.Add(1)

		go func() {
			serverRun.Done()
			server.RunServer()
		}()

		defer server.StopServer()

		serverRun.Wait()

		// Small sleep to allow the server to start up.
		time.Sleep(100 * time.Millisecond)

		return suite.Run()
	}()

	os.Exit(code)
}

func RunCMDTest(test *testing.T, mainFunc func(), args []string) (string, string, int, error) {
	// Suppress exits to capture exit codes.
	util.ShouldExitForTesting = false
	defer func() {
		util.ShouldExitForTesting = true
	}()

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

	exitCode := util.GetLastExitCode()

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

type CommonCMDTestCases struct {
	ExpectedExitCode int
	ExpectedStdout   string
	ExpectedStderr   string
}

func RunCommonCMDTests(test *testing.T, mainFunc func(), args []string, commonCases CommonCMDTestCases, prefix string) (string, string, int) {
	stdout, stderr, exitCode, err := RunCMDTest(test, mainFunc, args)
	if err != nil {
		test.Errorf("%sCMD run returned an error: '%v'.", prefix, err)
	}

	if len(stderr) > 0 && commonCases.ExpectedStderr == "" {
		test.Errorf("%sCMD has content in stderr: '%s'.", prefix, stderr)
	}

	if !strings.Contains(stderr, commonCases.ExpectedStderr) {
		test.Errorf("%sUnexpected stderr. Expected substring: '%s', Actual stderr: '%s'.", prefix, commonCases.ExpectedStderr, stderr)
	}

	if commonCases.ExpectedExitCode != exitCode {
		test.Errorf("%sUnexpected exit code. Expected: '%d', Actual: '%d'.", prefix, commonCases.ExpectedExitCode, exitCode)
	}

	if commonCases.ExpectedStdout != stdout {
		test.Errorf("%sUnexpected output. Expected:\n'%s',\nActual:\n'%s'.", prefix, commonCases.ExpectedStdout, stdout)
	}

	return stdout, stderr, exitCode
}

func RunVerboseCMDTests(test *testing.T, mainFunc func(), args []string, prefix string) {
	args = append(args, "--verbose")

	_, _, _, err := RunCMDTest(test, mainFunc, args)
	if err != nil {
		test.Errorf("%sCMD run returned an error when testing verbose: '%v'.", prefix, err)
	}
}
