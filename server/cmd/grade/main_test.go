package main

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const USERNAME = "_test_user_";

type TestSubmission struct {
    Course string `json:"course"`
    Assignment string `json:"assignment"`
    Result model.GradingResult `json:"result"`
}

func TestSubmissions(test *testing.T) {
    testsDir := filepath.Join(util.GetThisDir(), "..", "..", "..", "tests");

    tempDir, err := os.MkdirTemp("", "submission-tests-");
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not create temp dir.");
    }
    defer os.RemoveAll(tempDir);

    err = grader.LoadCoursesFromDir(testsDir);
    if (err != nil) {
        log.Fatal().Err(err).Msg("Could not load courses.");
    }

    for _, course := range grader.GetCourses() {
        course.Dir = filepath.Join(tempDir, course.ID);
    }

    testResults, err := util.FindFiles("test-submission.json", testsDir);
    if (err != nil) {
        log.Fatal().Err(err).Str("path", testsDir).Msg("Could not find test results.");
    }

    if (len(testResults) == 0) {
        test.Fatalf("Could not find any test cases in '%s'.", testsDir);
    }

    test.Parallel();

    for _, testResultPath := range testResults {
        testID := strings.TrimPrefix(testResultPath, testsDir);
        ok := test.Run(testID, func(test *testing.T) {
            var testSubmission TestSubmission;
            err := util.JSONFromFile(testResultPath, &testSubmission);
            if (err != nil) {
                test.Fatalf("Failed to load test submission: '%s': '%v'.", testResultPath, err);
            }

            assignment := grader.GetAssignment(testSubmission.Course, testSubmission.Assignment);
            if (assignment == nil) {
                test.Fatalf("Could not find assignment '%s' from course '%s'.", testSubmission.Assignment, testSubmission.Course);
            }

            result, err := assignment.Grade(filepath.Dir(testResultPath), USERNAME);
            if (err != nil) {
                test.Fatalf("Failed to grade assignment: '%v'.", err);
            }

            if (!result.Equals(&testSubmission.Result)) {
                test.Fatalf("Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.", result, &testSubmission.Result);
            }

        });

        if (!ok) {
            test.Fatalf("Failed to run test: '%s'.", testID);
        }
    }
}
