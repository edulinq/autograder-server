package proxy

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/grader"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestProxySubmitBase(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testSubmissions, err := grader.GetTestSubmissions(config.GetTestdataDir(), !config.DOCKER_DISABLE.Get())
	if err != nil {
		test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.GetTestdataDir(), err)
	}

	verifySuccessfulTestSubmissions(test, testSubmissions, nil)
}

func TestProxySubmitAtDueDate(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	testSubmissions, err := grader.GetTestSubmissions(config.GetTestdataDir(), !config.DOCKER_DISABLE.Get())
	if err != nil {
		test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.GetTestdataDir(), err)
	}

	testSubmission := testSubmissions[0]

	dueDate := timestamp.Timestamp(123456)
	testSubmission.Assignment.DueDate = &dueDate
	db.MustSaveAssignment(testSubmission.Assignment)

	verifySuccessfulTestSubmissions(test, []*grader.TestSubmissionInfo{testSubmission}, nil)
}

func TestProxySubmitProxyTime(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Disable testing mode to check for rejection.
	config.UNIT_TESTING_MODE.Set(false)
	defer config.UNIT_TESTING_MODE.Set(true)

	testSubmissions, err := grader.GetTestSubmissions(config.GetTestdataDir(), !config.DOCKER_DISABLE.Get())
	if err != nil {
		test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.GetTestdataDir(), err)
	}

	testSubmission := testSubmissions[0]
	proxyTime := timestamp.Timestamp(987654)

	verifySuccessfulTestSubmissions(test, []*grader.TestSubmissionInfo{testSubmission}, &proxyTime)
}

// Proxy submissions are not subject to submission restrictions, so we expect successful responses.
func TestRejectSubmissionMaxAttempts(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Disable testing mode to check for rejection.
	config.UNIT_TESTING_MODE.Set(false)
	defer config.UNIT_TESTING_MODE.Set(true)

	testSubmissions, err := grader.GetTestSubmissions(config.GetTestdataDir(), !config.DOCKER_DISABLE.Get())
	if err != nil {
		test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.GetTestdataDir(), err)
	}

	testSubmission := testSubmissions[0]

	course := db.MustGetCourse(testSubmission.Assignment.GetCourse().GetID())
	course.SubmissionLimit = &model.SubmissionLimitInfo{
		Max: util.IntPointer(0),
	}
	db.MustSaveCourse(course)

	verifySuccessfulTestSubmissions(test, []*grader.TestSubmissionInfo{testSubmission}, nil)
}

// Proxy submissions are not subject to submission restrictions, so we expect successful responses.
func TestRejectLateSubmission(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Disable testing mode to check for rejection.
	config.UNIT_TESTING_MODE.Set(false)
	defer config.UNIT_TESTING_MODE.Set(true)

	testSubmissions, err := grader.GetTestSubmissions(config.GetTestdataDir(), !config.DOCKER_DISABLE.Get())
	if err != nil {
		test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.GetTestdataDir(), err)
	}

	testSubmission := testSubmissions[0]

	// Assignment due at Unix Epoch.
	dueDate := timestamp.Zero()
	testSubmission.Assignment.DueDate = &dueDate
	db.MustSaveAssignment(testSubmission.Assignment)

	verifySuccessfulTestSubmissions(test, []*grader.TestSubmissionInfo{testSubmission}, nil)
}

func TestProxySubmitErrors(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Disable testing mode to check for rejection.
	config.UNIT_TESTING_MODE.Set(false)
	defer config.UNIT_TESTING_MODE.Set(true)

	testSubmissions, err := grader.GetTestSubmissions(config.GetTestdataDir(), !config.DOCKER_DISABLE.Get())
	if err != nil {
		test.Fatalf("Failed to get test submissions in '%s': '%v'.", config.GetTestdataDir(), err)
	}

	testSubmission := testSubmissions[0]

	testCases := []struct {
		proxyEmail        string
		userEmail         string
		expectedFoundUser bool
		permError         bool
	}{
		// Unknown proxy emails.
		{"zzz@test.edulinq.org", "course-admin", false, false},
		{"zzz@test.edulinq.org", "course-grader", false, false},

		// Permissions errors.
		{"course-student@test.edulinq.org", "course-student", false, true},
		{"zzz@test.edulinq.org", "course-other", false, true},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"course-id":     testSubmission.Assignment.GetCourse().GetID(),
			"assignment-id": testSubmission.Assignment.GetID(),
			"proxy-email":   testCase.proxyEmail,
			"proxy-time":    nil,
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/proxy/submit`, fields, testSubmission.Files, testCase.userEmail)
		if !response.Success {
			if testCase.permError {
				expectedLocator := "-020"
				if response.Locator != expectedLocator {
					test.Errorf("Case %d: Incorrect error returned. Expected '%s', found '%s'.",
						i, expectedLocator, response.Locator)
				}
			} else {
				test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response)
			}

			continue
		}

		if testCase.permError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		var responseContent SubmitResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if responseContent.FoundUser != testCase.expectedFoundUser {
			test.Errorf("Case %d: Unexpected found user. Expected: '%v', actual: '%v'.", i, testCase.expectedFoundUser, responseContent.FoundUser)
			continue
		}
	}
}

func verifySuccessfulTestSubmissions(test *testing.T, testSubmissions []*grader.TestSubmissionInfo, proxyTime *timestamp.Timestamp) {
	for i, testSubmission := range testSubmissions {
		fields := map[string]any{
			"course-id":     testSubmission.Assignment.GetCourse().GetID(),
			"assignment-id": testSubmission.Assignment.GetID(),
			"proxy-email":   "course-student@test.edulinq.org",
			"proxy-time":    proxyTime,
		}

		proxyTimeLowerBound := grader.ResolveProxyTime(proxyTime, testSubmission.Assignment)

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/proxy/submit`, fields, testSubmission.Files, "course-admin")
		if !response.Success {
			test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, util.MustToJSONIndent(response))
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

		startTime := responseContent.GradingInfo.GradingStartTime
		proxyTimeUpperBound := grader.ResolveProxyTime(proxyTime, testSubmission.Assignment)

		if startTime < *proxyTimeLowerBound || startTime > *proxyTimeUpperBound {
			test.Errorf("Case %d: Unexpected grading start time. Expected a start time in the range: ['%d', '%d'], actual: '%d'.",
				i, proxyTimeLowerBound, proxyTimeUpperBound, startTime)
			continue
		}

		if startTime > responseContent.GradingInfo.GradingEndTime {
			test.Errorf("Case %d: Unexpected grading end time. Expected a time after: '%d', actual: '%d'.",
				i, startTime, responseContent.GradingInfo.GradingEndTime)
			continue
		}
	}
}
