package grader

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const ENV_TESTS_DIR = "AUTOGRADER__TEST__TESTS_DIR";
const BASE_TEST_USER = "test_user@test.com";

func TestDockerSubmissions(test *testing.T) {
    if (!CanAccessDocker()) {
        test.Fatal("Could not access docker.");
    }

    runSubmissionTests(test, true, true);
}

func TestNoDockerSubmissions(test *testing.T) {
    runSubmissionTests(test, true, false);
}

func runSubmissionTests(test *testing.T, parallel bool, docker bool) {
    testsDir := os.Getenv(ENV_TESTS_DIR);
    if (testsDir == "") {
        test.Fatalf("No tests dir set (env variable '%s').", ENV_TESTS_DIR);
    }

    err := LoadCoursesFromDir(testsDir);
    if (err != nil) {
        test.Fatalf("Could not load courses from '%s': '%v'.", testsDir, err);
    }

    _, err = BuildDockerImagesJoinErrors(NewDockerBuildOptions());
    if (err != nil) {
        test.Fatalf("Failed to build docker images: '%v'.", err);
    }

    tempDir, err := os.MkdirTemp("", "submission-tests-");
    if (err != nil) {
        test.Fatalf("Could not create temp dir: '%v'.", err);
    }
    defer os.RemoveAll(tempDir);

    testSubmissionPaths, err := util.FindFiles("test-submission.json", testsDir);
    if (err != nil) {
        test.Fatalf("Could not find test results in '%s': '%v'.", testsDir, err);
    }

    if (len(testSubmissionPaths) == 0) {
        test.Fatalf("Could not find any test cases in '%s'.", testsDir);
    }

    gradeOptions := GradeOptions{
        UseFakeSubmissionsDir: true,
        NoDocker: !docker,
    };

    failedTests := make([]string, 0);

    for i, testSubmissionPath := range testSubmissionPaths {
        testID := strings.TrimPrefix(testSubmissionPath, testsDir);
        user := fmt.Sprintf("%03d_%s", i, BASE_TEST_USER);

        ok := test.Run(testID, func(test *testing.T) {
            if (parallel) {
                test.Parallel();
            }

            var testSubmission model.TestSubmission;
            err := util.JSONFromFile(testSubmissionPath, &testSubmission);
            if (err != nil) {
                test.Fatalf("Failed to load test submission: '%s': '%v'.", testSubmissionPath, err);
            }

            assignment := fetchTestSubmissionAssignment(testSubmissionPath);
            if (assignment == nil) {
                test.Fatalf("Could not find assignment for test submission '%s'.", testSubmissionPath);
            }

            result, err := Grade(assignment, filepath.Dir(testSubmissionPath), user, gradeOptions);
            if (err != nil) {
                test.Fatalf("Failed to grade assignment: '%v'.", err);
            }

            if (!result.Equals(testSubmission.Result, !testSubmission.IgnoreMessages)) {
                test.Fatalf("Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.", result, &testSubmission.Result);
            }

        });

        if (!ok) {
            failedTests = append(failedTests, testID);
        }
    }

    if (len(failedTests) > 0) {
        test.Fatalf("Failed to run submission test(s): '%s'.", failedTests);
    }
}

// Test submission are withing their assignment's directory,
// just check the source dirs for existing courses and assignments.
func fetchTestSubmissionAssignment(testSubmissionPath string) *model.Assignment {
    testSubmissionPath = util.MustAbs(testSubmissionPath);

    for _, course := range GetCourses() {
        if (!util.PathHasParent(testSubmissionPath, filepath.Dir(course.SourcePath))) {
            continue;
        }

        for _, assignment := range course.Assignments {
            if (util.PathHasParent(testSubmissionPath, filepath.Dir(assignment.SourcePath))) {
                return assignment;
            }
        }
    }

    return nil;
}
