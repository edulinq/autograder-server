package submissions

import (
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
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
			// Ensure grading passes even if it is past the due date.
			// Late submissions are tested thoroughly in TestRejectLateSubmission.
			"allow-late": true,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/submit`, fields, testSubmission.Files, "course-student")
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			continue
		}

		var responseContent SubmitResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if !responseContent.GradingSuccess {
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

	course := db.MustGetCourse("course101")
	course.SubmissionLimit = &model.SubmissionLimitInfo{
		Max: util.IntPointer(0),
	}
	db.MustSaveCourse(course)

	assignment := db.MustGetTestAssignment()
	paths := []string{filepath.Join(assignment.GetSourceDir(), SUBMISSION_RELPATH)}

	fields := map[string]any{
		"course-id":     "course101",
		"assignment-id": "hw0",
	}

	response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/submit`, fields, paths, "course-student")
	if !response.Success {
		test.Fatalf("Response is not a success when it should be: '%v'.", response)
	}

	var responseContent SubmitResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

	if responseContent.GradingSuccess {
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

func TestRejectLateSubmission(test *testing.T) {
	defer db.ResetForTesting()

	assignment := db.MustGetTestAssignment()
	paths := []string{filepath.Join(assignment.GetSourceDir(), SUBMISSION_RELPATH)}

	timeDeltaPattern := regexp.MustCompile(`(\d+h)?(\d+m)?(\d+\.)?\d+[mun]?s`)
	timeDeltaReplacement := "<time-delta:TIME>"

	testCases := []struct {
		allowLate    bool
		dueDate      timestamp.Timestamp
		expectReject bool
	}{
		// Assignment due at Unix Epoch, reject without allow late.
		{
			allowLate:    true,
			dueDate:      timestamp.Zero(),
			expectReject: false,
		},
		{
			allowLate:    false,
			dueDate:      timestamp.Zero(),
			expectReject: true,
		},

		// Assignment due tomorrow, will never reject.
		{
			allowLate:    true,
			dueDate:      timestamp.FromGoTime(time.Now().Add(24 * time.Hour)),
			expectReject: false,
		},
		{
			allowLate:    false,
			dueDate:      timestamp.FromGoTime(time.Now().Add(24 * time.Hour)),
			expectReject: false,
		},
	}

	for i, testCase := range testCases {
		db.ResetForTesting()

		assignment.DueDate = &testCase.dueDate
		db.MustSaveAssignment(assignment)

		fields := map[string]any{
			"course-id":     assignment.GetCourse().GetID(),
			"assignment-id": assignment.GetID(),
			"allow-late":    testCase.allowLate,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/submit`, fields, paths, "course-student")
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.",
				i, util.MustToJSONIndent(response))
			continue
		}

		var responseContent SubmitResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		// If grading succeeds when we expect it to be rejected (which sets grading success to false).
		if responseContent.GradingSuccess == testCase.expectReject {
			test.Errorf("Case %d: Unexpected grading success result. Expected: '%v', actual: '%v': Full content: '%v'.",
				i, responseContent.GradingSuccess, testCase.expectReject, util.MustToJSONIndent(responseContent))
			continue
		}

		if responseContent.Rejected != testCase.expectReject {
			test.Errorf("Case %d: Unexpected response rejection status. Expected: '%v', actual: '%v'. Full content: '%v'.",
				i, responseContent.Rejected, testCase.expectReject, util.MustToJSONIndent(responseContent))
			continue
		}

		if testCase.expectReject {
			expected := (&grader.RejectLate{assignment.Name, *assignment.DueDate}).String()
			expected = timeDeltaPattern.ReplaceAllString(expected, timeDeltaReplacement)
			actual := timeDeltaPattern.ReplaceAllString(responseContent.Message, timeDeltaReplacement)
			if expected != actual {
				test.Fatalf("Case %d: Did not get the expected rejection reason. Expected: '%s', Actual: '%s'.",
					i, expected, actual)
			}
		} else {
			if responseContent.Message != "" {
				test.Errorf("Case %d: Response has a reject reason when it should not: '%v'.",
					i, util.MustToJSONIndent(responseContent))
				continue
			}
		}
	}
}
