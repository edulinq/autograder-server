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

// Proxy resubmissions should never be rejected.
func TestProxyResubmit(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	// Disable testing mode to check for rejection.
	config.UNIT_TESTING_MODE.Set(false)
	defer config.UNIT_TESTING_MODE.Set(true)

	// Note that computation of these paths is deferred until test time.
	studentGradingResults := map[string]*model.GradingResult{
		"1697406256": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406256")),
		"1697406265": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406265")),
		"1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
		"course101::hw0::student@test.edulinq.org::1697406256": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406256")),
		"course101::hw0::student@test.edulinq.org::1697406265": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406265")),
		"course101::hw0::student@test.edulinq.org::1697406272": model.MustLoadGradingResult(getTestSubmissionResultPath("1697406272")),
	}

	assignment := db.MustGetTestAssignment()

	testCases := []struct {
		proxyEmail              string
		userEmail               string
		targetSubmission        string
		proxyTime               *timestamp.Timestamp
		expectedFoundUser       bool
		expectedFoundSubmission bool
		permError               bool
	}{
		// Valid proxy resubmissions.
		// Short ID, nil proxy time.
		{"course-student@test.edulinq.org", "course-admin", "1697406256", nil, true, true, false},
		{"course-student@test.edulinq.org", "course-admin", "1697406265", nil, true, true, false},
		{"course-student@test.edulinq.org", "course-admin", "1697406272", nil, true, true, false},

		// Short ID, explicit proxy time.
		{"course-student@test.edulinq.org", "course-admin", "1697406256", getTestProxyTime(), true, true, false},
		{"course-student@test.edulinq.org", "course-admin", "1697406265", getTestProxyTime(), true, true, false},
		{"course-student@test.edulinq.org", "course-admin", "1697406272", getTestProxyTime(), true, true, false},

		// Long ID, nil proxy time.
		{"course-student@test.edulinq.org", "course-admin", "course101::hw0::student@test.edulinq.org::1697406256", nil, true, true, false},
		{"course-student@test.edulinq.org", "course-admin", "course101::hw0::student@test.edulinq.org::1697406265", nil, true, true, false},
		{"course-student@test.edulinq.org", "course-admin", "course101::hw0::student@test.edulinq.org::1697406272", nil, true, true, false},

		// Long ID, explicit proxy time.
		{"course-student@test.edulinq.org", "course-admin", "course101::hw0::student@test.edulinq.org::1697406256", getTestProxyTime(), true, true, false},
		{"course-student@test.edulinq.org", "course-admin", "course101::hw0::student@test.edulinq.org::1697406265", getTestProxyTime(), true, true, false},
		{"course-student@test.edulinq.org", "course-admin", "course101::hw0::student@test.edulinq.org::1697406272", getTestProxyTime(), true, true, false},

		// Invalid proxy resubmissions.
		// Unknown proxy emails.
		{"zzz@test.edulinq.org", "course-admin", "", nil, false, false, false},
		{"zzz@test.edulinq.org", "course-grader", "", nil, false, false, false},

		// Unknown target submissions.
		{"course-student@test.edulinq.org", "course-admin", "ZZZ", nil, true, false, false},
		{"course-student@test.edulinq.org", "course-grader", "ZZZ", nil, true, false, false},

		// Permissions errors.
		{"course-student@test.edulinq.org", "course-student", "", nil, false, false, true},
		{"zzz@test.edulinq.org", "course-other", "", nil, false, false, true},
	}

	for i, testCase := range testCases {
		fields := map[string]any{
			"course-id":         assignment.GetCourse().GetID(),
			"assignment-id":     assignment.GetID(),
			"target-submission": testCase.targetSubmission,
			"proxy-email":       testCase.proxyEmail,
			"proxy-time":        testCase.proxyTime,
		}

		proxyTimeLowerBound := grader.ResolveProxyTime(testCase.proxyTime, assignment)
		// If we are given a proxy time, make sure it is unchanged.
		if testCase.proxyTime != nil && *testCase.proxyTime != *proxyTimeLowerBound {
			test.Errorf("Case %d: Unexpected proxy time lower bound. Expected: '%d', actual: '%d'.",
				i, *testCase.proxyTime, *proxyTimeLowerBound)
			continue
		}

		response := core.SendTestAPIRequestFull(test, `courses/assignments/submissions/proxy/resubmit`, fields, nil, testCase.userEmail)
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

		var responseContent ResubmitResponse
		util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent)

		if responseContent.FoundUser != testCase.expectedFoundUser {
			test.Errorf("Case %d: Unexpected found user. Expected: '%v', actual: '%v'.", i, testCase.expectedFoundUser, responseContent.FoundUser)
			continue
		}

		if responseContent.FoundSubmission != testCase.expectedFoundSubmission {
			test.Errorf("Case %d: Unexpected found submission. Expected: '%v', actual: '%v'.", i, testCase.expectedFoundSubmission, responseContent.FoundSubmission)
			continue
		}

		if testCase.expectedFoundUser == false || testCase.expectedFoundSubmission == false {
			if responseContent.GradingSuccess {
				test.Errorf("Case %d: Response is a grading success when it not should be: '%v'.", i, responseContent)
			}

			continue
		}

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

		expectedGradingResult := studentGradingResults[testCase.targetSubmission]
		if !responseContent.GradingInfo.Equals(*expectedGradingResult.Info, false) {
			test.Errorf("Case %d: Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.",
				i, responseContent.GradingInfo, expectedGradingResult)
			continue
		}

		// Fetch the most recent submission from the DB and ensure it matches.
		submission, err := db.GetSubmissionResult(assignment, testCase.proxyEmail, "")
		if err != nil {
			test.Errorf("Case %d: Failed to get submission: '%v'.", i, err)
			continue
		}

		if !responseContent.GradingInfo.Equals(*submission, false) {
			test.Errorf("Case %d: Actual output:\n---\n%v\n---\ndoes not match database value:\n---\n%v\n---\n.",
				i, responseContent.GradingInfo, submission)
			continue
		}

		proxyTimeUpperBound := grader.ResolveProxyTime(testCase.proxyTime, assignment)
		// If we are given a proxy time, make sure it is unchanged.
		if testCase.proxyTime != nil && *testCase.proxyTime != *proxyTimeUpperBound {
			test.Errorf("Case %d: Unexpected proxy time upper bound. Expected: '%d', actual: '%d'.",
				i, *testCase.proxyTime, *proxyTimeUpperBound)
			continue
		}

		startTime := responseContent.GradingInfo.GradingStartTime

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

		// If we resolved the proxy time, ensure it is not late.
		if testCase.proxyTime == nil && assignment.DueDate != nil && startTime > *assignment.DueDate {
			test.Errorf("Case %d: A submission with a resolved proxy time marked late. Due date: '%d', proxy time: '%d'.",
				i, *assignment.DueDate, startTime)
			continue
		}
	}
}

func getTestProxyTime() *timestamp.Timestamp {
	proxyTime := timestamp.Timestamp(987654)
	return &proxyTime
}
