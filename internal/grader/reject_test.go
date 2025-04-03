package grader

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

const TEST_COURSE_ID = "course-languages"
const TEST_ASSIGNMENT_ID = "bash"

var SUBMISSION_RELPATH string = filepath.Join("test-submissions", "solution")

// Ensure that admin submissions are never rejected.
func TestRejectSubmissionAdminOverride(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestSubmissionAssignment()
	assignment.DueDate = nil

	// Set the max submissions to zero.
	maxValue := 0
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue}

	// Make a submission that should be rejected, but is not.
	submitForRejection(test, assignment, "server-admin@test.edulinq.org", false, nil)
}

func TestRejectSubmissionMaxAttempts(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestSubmissionAssignment()
	assignment.DueDate = nil

	// Set the max submissions to zero.
	maxValue := 0
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue}

	// Make a submission that should be rejected.
	submitForRejection(test, assignment, "course-other@test.edulinq.org", false, &RejectMaxAttempts{0})
}

func TestRejectSubmissionMaxAttemptsInfinite(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestSubmissionAssignment()
	assignment.DueDate = nil

	// Set the max submissions to empty (infinite).
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{}

	// All submissions should pass.
	submitForRejection(test, assignment, "course-other@test.edulinq.org", false, nil)

	// Set the max submissions to nagative (infinite).
	maxValue := -1
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue}

	// All submissions should pass.
	submitForRejection(test, assignment, "course-other@test.edulinq.org", false, nil)
}

func TestRejectSubmissionMaxWindowAttempts(test *testing.T) {
	testMaxWindowAttempts(test, "course-other@test.edulinq.org", true)
}

func TestRejectSubmissionMaxWindowAttemptsAdmin(test *testing.T) {
	testMaxWindowAttempts(test, "course-grader@test.edulinq.org", false)
}

func TestRejectWindowMaxMessage(test *testing.T) {
	now := timestamp.Timestamp(0)

	testCases := []struct {
		input    RejectWindowMax
		expected string
	}{
		{
			RejectWindowMax{1, util.DurationSpec{Hours: 1}, timestamp.Timestamp(0)},
			"Reached the number of max attempts (1) within submission window (every 1 hours). Next allowed submission time is <timestamp:3600000> (in 1h0m0s).",
		},
		{
			RejectWindowMax{1, util.DurationSpec{Days: 1, Hours: 1}, timestamp.Timestamp(0)},
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

func TestRejectLateSubmissionWithoutAllow(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestSubmissionAssignment()

	// Set a dummy submission limit.
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{}

	// Set the due date to be the Unix epoch.
	timestamp := timestamp.Zero()
	assignment.DueDate = &timestamp

	submitForRejection(test, assignment, "course-other@test.edulinq.org", false, &RejectLate{assignment.Name, *assignment.DueDate})
}

func TestRejectLateSubmissionWithAllow(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestSubmissionAssignment()

	// Set a dummy submission limit.
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{}

	// Set the due date to be the Unix epoch.
	timestamp := timestamp.Zero()
	assignment.DueDate = &timestamp

	submitForRejection(test, assignment, "course-other@test.edulinq.org", true, nil)
}

func testMaxWindowAttempts(test *testing.T, user string, expectReject bool) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	assignment := db.MustGetTestSubmissionAssignment()
	assignment.DueDate = nil
	duration := util.DurationSpec{Days: 1000}

	// Set the submission limit window to 1 attempt in a large duration.
	assignment.SubmissionLimit = &model.SubmissionLimitInfo{
		Window: &model.SubmittionLimitWindow{
			AllowedAttempts: 1,
			Duration:        duration,
		},
	}

	// Make a submission that should pass.
	result, _, _ := submitForRejection(test, assignment, user, false, nil)

	// Make a submission that should be rejected.
	var reason RejectReason
	if expectReject {
		reason = &RejectWindowMax{1, duration, result.Info.GradingStartTime}
	}

	submitForRejection(test, assignment, user, false, reason)
}

func submitForRejection(test *testing.T, assignment *model.Assignment, user string, allowLate bool, expectedRejection RejectReason) (
	*model.GradingResult, RejectReason, error) {
	// Disable testing mode to check for rejection.
	config.UNIT_TESTING_MODE.Set(false)
	defer config.UNIT_TESTING_MODE.Set(true)

	submissionPath := filepath.Join(assignment.GetSourceDir(), SUBMISSION_RELPATH)

	err := assignment.SubmissionLimit.Validate()
	if err != nil {
		test.Fatalf("Failed to validate submission limit: '%v'.", err)
	}

	gradeOptions := GetDefaultGradeOptions()
	gradeOptions.AllowLate = allowLate

	result, reject, softError, err := Grade(context.Background(), assignment, submissionPath, user, TEST_MESSAGE, gradeOptions)
	if err != nil {
		test.Fatalf("Failed to grade assignment: '%v'.", err)
	}

	if softError != "" {
		test.Fatalf("Submission got a soft error: '%s'.", softError)
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
