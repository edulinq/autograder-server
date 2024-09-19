package grader

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

var SUBMISSION_RELPATH string = filepath.Join("test-submissions", "solution")

func TestRejectSubmissionMaxAttempts(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestAssignment()

	// Set the max submissions to zero.
	maxValue := 0
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue}

	// Make a submission that should be rejected.
	submitForRejection(test, assignment, "course-other@test.edulinq.org", &RejectMaxAttempts{0})
}

func TestRejectSubmissionMaxAttemptsInfinite(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestAssignment()

	// Set the max submissions to empty (infinite).
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{}

	// All submissions should pass.
	submitForRejection(test, assignment, "course-other@test.edulinq.org", nil)

	// Set the max submissions to nagative (infinite).
	maxValue := -1
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue}

	// All submissions should pass.
	submitForRejection(test, assignment, "course-other@test.edulinq.org", nil)
}

func TestRejectSubmissionMaxWindowAttempts(test *testing.T) {
	testMaxWindowAttemps(test, "course-other@test.edulinq.org", true)
}

func TestRejectSubmissionMaxWindowAttemptsAdmin(test *testing.T) {
	testMaxWindowAttemps(test, "course-grader@test.edulinq.org", false)
}

func TestRejectWindowMaxMessage(test *testing.T) {
	now := timestamp.Timestamp(0)

	testCases := []struct {
		input    RejectWindowMax
		expected string
	}{
		{
			RejectWindowMax{1, common.DurationSpec{Hours: 1}, timestamp.Timestamp(0)},
			"Reached the number of max attempts (1) within submission window (every 1 hours). Next allowed submission time is <timestamp:3600000> (in 1h0m0s).",
		},
		{
			RejectWindowMax{1, common.DurationSpec{Days: 1, Hours: 1}, timestamp.Timestamp(0)},
			"Reached the number of max attempts (1) within submission window (every 1 days, 1 hours). Next allowed submission time is <timestamp:90000000> (in 25h0m0s).",
		},
	}

	for i, testCase := range testCases {
		actual := testCase.input.fullString(now)
		if testCase.expected != actual {
			test.Errorf("Case %d: Message does not match. Expected: '%s', Actual: '%s'.", i, testCase.expected, actual)
			continue
		}
	}
}

func testMaxWindowAttemps(test *testing.T, user string, expectReject bool) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestAssignment()
	duration := common.DurationSpec{Days: 1000}

	// Set the submission limit window to 1 attempt in a large duration.
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{
		Window: &model.SubmittionLimitWindow{
			AllowedAttempts: 1,
			Duration:        duration,
		},
	}

	// Make a submission that should pass.
	result, _, _ := submitForRejection(test, assignment, user, nil)

	// Make a submission that should be rejected.
	var reason RejectReason
	if expectReject {
		reason = &RejectWindowMax{1, duration, result.Info.GradingStartTime}
	}

	submitForRejection(test, assignment, user, reason)
}

func submitForRejection(test *testing.T, assignment *model.Assignment, user string, expectedRejection RejectReason) (
	*model.GradingResult, RejectReason, error) {
	// Disable testing mode to check for rejection.
	config.TESTING_MODE.Set(false)
	defer config.TESTING_MODE.Set(true)

	submissionPath := filepath.Join(assignment.GetSourceDir(), SUBMISSION_RELPATH)

	err := assignment.SubmissionLimit.Validate()
	if err != nil {
		test.Fatalf("Failed to validate submission limit: '%v'.", err)
	}

	result, reject, err := GradeDefault(assignment, submissionPath, user, TEST_MESSAGE)
	if err != nil {
		test.Fatalf("Failed to grade assignment: '%v'.", err)
	}

	if expectedRejection == nil {
		// Submission should go through.

		if reject != nil {
			test.Fatalf("Submission was rejected: '%s'.", reject.String())
		}

		if result == nil {
			test.Fatalf("Did not get a grading result.")
		}
	} else {
		// Submission should be rejected.

		if result != nil {
			test.Fatalf("Should not get a grading result.")
		}

		if reject == nil {
			test.Fatalf("Submission was not rejected when it should have been.")
		}

		if !reflect.DeepEqual(expectedRejection, reject) {
			test.Fatalf("Did not get the expected rejection. Expected: '%+v', Actual: '%+v'.", expectedRejection, reject)
		}
	}

	return result, reject, err
}
