package submissions

import (
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/util"
)

var SUBMISSION_RELPATH string = filepath.Join("test-submissions", "solution", "submission.py")

func TestSubmit(test *testing.T) {
	testSubmissions, err := grader.GetTestSubmissions(config.GetTestdataDir(), !config.DOCKER_DISABLE.Get())
	if err != nil {
		test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.GetTestdataDir(), err)
	}

	for i, testSubmission := range testSubmissions {
		fields := map[string]any{
			"course-id":     testSubmission.Assignment.GetCourse().GetID(),
			"assignment-id": testSubmission.Assignment.GetID(),
		}

		response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/assignments/submissions/submit`), fields, testSubmission.Files, "course-student")
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			continue
		}

		var responseContent SubmitResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !responseContent.GradingSucess {
			test.Errorf("Case %d: Response is not a grading success when it should be: '%v'.", i, responseContent)
			continue
		}

		if responseContent.Rejected {
			test.Errorf("Case %d: Response is rejected when it should not be: '%v'.", i, responseContent)
			continue
		}

		if responseContent.Message != "" {
			test.Errorf("Case %d: Response has a reject reason when it should not: '%v'.", i, responseContent)
			continue
		}

		if !responseContent.GradingInfo.Equals(*testSubmission.TestSubmission.GradingInfo, !testSubmission.TestSubmission.IgnoreMessages) {
			test.Errorf("Case %d: Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.",
				i, responseContent.GradingInfo, testSubmission.TestSubmission.GradingInfo)
			continue
		}

		// Fetch the most recent submission from the DB and ensure it matches.
		submission, err := db.GetSubmissionResult(testSubmission.Assignment, "course-student@test.edulinq.org", "")
		if err != nil {
			test.Errorf("Case %d: Failed to get submission: '%v'.", i, err)
			continue
		}

		if !responseContent.GradingInfo.Equals(*submission, !testSubmission.TestSubmission.IgnoreMessages) {
			test.Errorf("Case %d: Actual output:\n---\n%v\n---\ndoes not match database value:\n---\n%v\n---\n.",
				i, responseContent.GradingInfo, submission)
			continue
		}
	}
}

func TestRejectSubmissionMaxAttempts(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Disable testing mode to check for rejection.
	config.UNIT_TESTING_MODE.Set(false)
	defer config.UNIT_TESTING_MODE.Set(true)

	// Note that we are using a submission from a different assignment.
	assignment := db.MustGetTestAssignment()
	paths := []string{filepath.Join(assignment.GetSourceDir(), SUBMISSION_RELPATH)}

	fields := map[string]any{
		"course-id":     "course101-with-zero-limit",
		"assignment-id": "hw0",
	}

	response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`courses/assignments/submissions/submit`), fields, paths, "course-student")
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent SubmitResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if responseContent.GradingSucess {
		test.Fatalf("Response is a grading success when it should not be: '%v'.", responseContent)
	}

	if !responseContent.Rejected {
		test.Fatalf("Response is not rejected when it should be: '%v'.", responseContent)
	}

	if responseContent.Message == "" {
		test.Fatalf("Response does not have a reject reason when it should: '%v'.", responseContent)
	}

	expected := (&grader.RejectMaxAttempts{0}).String()
	if expected != responseContent.Message {
		test.Fatalf("Did not get the expected rejection reason. Expected: '%s', Actual: '%s'.",
			expected, responseContent.Message)
	}
}
