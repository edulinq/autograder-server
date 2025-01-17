package grader

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/util"
)

const BASE_TEST_USER = "course-student@test.edulinq.org"
const TEST_MESSAGE = ""

func TestDockerSubmissions(test *testing.T) {
	if config.DOCKER_DISABLE.Get() {
		test.Skip("Docker is disabled, skipping test.")
	}

	if !docker.CanAccessDocker() {
		test.Fatal("Could not access docker.")
	}

	runSubmissionTests(test, false, true)
}

func TestNoDockerSubmissions(test *testing.T) {
	oldDockerVal := config.DOCKER_DISABLE.Get()
	config.DOCKER_DISABLE.Set(true)
	defer config.DOCKER_DISABLE.Set(oldDockerVal)

	runSubmissionTests(test, false, false)
}

func runSubmissionTests(test *testing.T, parallel bool, useDocker bool) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Directory where all the test courses and other materials are located.
	baseDir := config.GetTestdataDir()

	if useDocker {
		for _, course := range db.MustGetCourses() {
			for _, assignment := range course.GetAssignments() {
				err := docker.BuildImageFromSource(assignment, false, false, docker.NewBuildOptions())
				if err != nil {
					test.Fatalf("Failed to build image '%s': '%v'.", assignment.FullID(), err)
				}
			}
		}
	}

	gradeOptions := GradeOptions{
		NoDocker: !useDocker,
	}

	testSubmissions, err := GetTestSubmissions(baseDir, useDocker)
	if err != nil {
		test.Fatalf("Error getting test submissions in '%s': '%v'.", baseDir, err)
	}

	if len(testSubmissions) == 0 {
		test.Fatalf("Could not find any test submissions in '%s'.", baseDir)
	}

	failedTests := make([]string, 0)

	for i, testSubmission := range testSubmissions {
		user := fmt.Sprintf("%03d_%s", i, BASE_TEST_USER)

		ok := test.Run(testSubmission.ID, func(test *testing.T) {
			if parallel {
				test.Parallel()
			}

			result, reject, softError, err := Grade(testSubmission.Assignment, testSubmission.Dir, user, TEST_MESSAGE, false, gradeOptions)
			if err != nil {
				if result != nil {
					fmt.Println("--- stdout ---")
					fmt.Println(result.Stdout)
					fmt.Println("--------------")

					fmt.Println("--- stderr ---")
					fmt.Println(result.Stderr)
					fmt.Println("--------------")
				}

				test.Fatalf("Failed to grade assignment: '%v'.", err)
			}

			if reject != nil {
				test.Fatalf("Submission was rejected: '%s'.", reject.String())
			}

			if softError != "" {
				test.Fatalf("Submission got a soft error: '%s'.", softError)
			}

			if !result.Info.Equals(*testSubmission.TestSubmission.GradingInfo, !testSubmission.TestSubmission.IgnoreMessages) {
				test.Fatalf("Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.",
					result.Info, testSubmission.TestSubmission.GradingInfo)
			}

		})

		if !ok {
			failedTests = append(failedTests, testSubmission.ID)
		}
	}

	if len(failedTests) > 0 {
		test.Fatalf("Failed to run submission test(s): '%s'.", failedTests)
	}
}

func TestGradeTimeoutDocker(test *testing.T) {
	if config.DOCKER_DISABLE.Get() {
		test.Skip("Docker is disabled, skipping test.")
	}

	if !docker.CanAccessDocker() {
		test.Fatal("Could not access docker.")
	}

	testGradeTimeout(test, false)
}

func TestGradeTimeoutNoDocker(test *testing.T) {
	oldValue := noDockerTimeoutWaitDelayMS
	noDockerTimeoutWaitDelayMS = 10
	defer func() {
		noDockerTimeoutWaitDelayMS = oldValue
	}()

	testGradeTimeout(test, true)
}

func testGradeTimeout(test *testing.T, noDocker bool) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	submissionDir := filepath.Join(util.ShouldGetThisDir(), "testdata", "bash-sleep")

	assignment := db.MustGetAssignment("course-languages", "bash")

	// Set a short timeout, which should ensure this submission runs out of time.
	assignment.MaxRuntimeSecs = 1

	options := GetDefaultGradeOptions()
	options.NoDocker = noDocker

	result, reject, softError, err := Grade(assignment, submissionDir, "course-student@test.edulinq.org", "", false, options)
	if err != nil {
		if result != nil {
			fmt.Println("--- stdout ---")
			fmt.Println(result.Stdout)
			fmt.Println("--------------")

			fmt.Println("--- stderr ---")
			fmt.Println(result.Stderr)
			fmt.Println("--------------")
		}

		test.Fatalf("Failed to grade assignment: '%v'.", err)
	}

	if reject != nil {
		test.Fatalf("Submission was rejected: '%s'.", reject.String())
	}

	if softError == "" {
		test.Fatalf("Submission did not get a soft error.")
	}

	expectedSubstring := "Submission has ran for too long and was killed."
	if !strings.Contains(softError, expectedSubstring) {
		test.Fatalf("Submission did not get the correct soft error. Expected substring: '%s', Actual string: '%s'.", expectedSubstring, softError)
	}
}

func TestGradeTruncatedOutputNoTruncation(test *testing.T) {
	testTruncatedOutput(test, 20, false)
}

func TestGradeTruncatedOutputTruncation(test *testing.T) {
	testTruncatedOutput(test, 1, true)
}

func testTruncatedOutput(test *testing.T, sizeKB int, expectedTruncated bool) {
	if config.DOCKER_DISABLE.Get() {
		test.Skip("Docker is disabled, skipping test.")
	}

	if !docker.CanAccessDocker() {
		test.Fatal("Could not access docker.")
	}

	db.ResetForTesting()
	defer db.ResetForTesting()

	defer config.DOCKER_MAX_OUTPUT_SIZE_KB.Set(config.DOCKER_MAX_OUTPUT_SIZE_KB.Get())
	config.DOCKER_MAX_OUTPUT_SIZE_KB.Set(sizeKB)

	submissionDir := filepath.Join(util.ShouldGetThisDir(), "testdata", "bash-outputsize")

	assignment := db.MustGetAssignment("course-languages", "bash")

	result, reject, softError, err := Grade(assignment, submissionDir, "course-student@test.edulinq.org", "", false, GradeOptions{})
	if err != nil {
		test.Fatalf("Failed to grade assignment: '%v'.", err)
	}

	if reject != nil {
		test.Fatalf("Submission was rejected: '%s'.", reject.String())
	}

	if softError != "" {
		test.Fatalf("Submission got a soft error: '%s'.", softError)
	}

	expectedSubstring := "Combined output (stdout + stderr) exceeds maximum size"

	if expectedTruncated {
		if !strings.Contains(result.Stdout, expectedSubstring) {
			test.Fatalf("Output was not truncated when it should have been.")
		}
	} else {
		if strings.Contains(result.Stdout, expectedSubstring) {
			test.Fatalf("Output was truncated when it should not have been.")
		}
	}
}
