package grader

import (
	"fmt"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
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

			result, reject, err := Grade(testSubmission.Assignment, testSubmission.Dir, user, TEST_MESSAGE, false, gradeOptions)
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
