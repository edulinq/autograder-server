package grader

import (
    "path/filepath"
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
)

var SUBMISSION_RELPATH string = filepath.Join("test-submissions", "solution");

func TestRejectSubmissionMaxAttempts(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();

    // Set the max submissions to zero.
    maxValue := 0
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue};

    // Make a submission that should be rejected.
    submitForRejection(test, assignment, &RejectMaxAttempts{0});
}

func TestRejectSubmissionMaxAttemptsInfinite(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();

    // Set the max submissions to empty (infinite).
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{};

    // All submissions should pass.
    submitForRejection(test, assignment, nil);

    // Set the max submissions to nagative (infinite).
    maxValue := -1
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{Max: &maxValue};

    // All submissions should pass.
    submitForRejection(test, assignment, nil);
}

func TestRejectSubmissionMaxWindowAttempts(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    assignment := db.MustGetTestAssignment();
    duration := common.DurationSpec{Days: 1000};

    // Set the submission limit window to 1 attempt in a large duration.
    assignment.SubmissionLimit = &model.SubmissionLimitInfo{
        Window: &model.SubmittionLimitWindow{
            AllowedAttempts: 1,
            Duration: duration,
        },
    };

    // Make a submission that should pass.
    result, _, _ := submitForRejection(test, assignment, nil);

    expectedTime, err := result.Info.GradingStartTime.Time();
    if (err != nil) {
        test.Fatalf("Failed to parse expected time: '%v'.", err);
    }

    // Make a submission that should be rejected.
    submitForRejection(test, assignment, &RejectWindowMax{1, duration, expectedTime});
}

func submitForRejection(test *testing.T, assignment *model.Assignment, expectedRejection RejectReason) (
        *model.GradingResult, RejectReason, error) {
    // Disable testing mode to check for rejection.
    config.TESTING_MODE.Set(false);
    defer config.TESTING_MODE.Set(true);

    submissionPath := filepath.Join(assignment.GetSourceDir(), SUBMISSION_RELPATH);

    err := assignment.SubmissionLimit.Validate();
    if (err != nil) {
        test.Fatalf("Failed to validate submission limit: '%v'.", err);
    }

    result, reject, err := GradeDefault(assignment, submissionPath, BASE_TEST_USER, TEST_MESSAGE);
    if (err != nil) {
        test.Fatalf("Failed to grade assignment: '%v'.", err);
    }

    if (expectedRejection == nil) {
        // Submission should go through.

        if (reject != nil) {
            test.Fatalf("Submission was rejected: '%s'.", reject.String());
        }

        if (result == nil) {
            test.Fatalf("Did not get a grading result.");
        }
    } else {
        // Submission should be rejected.

        if (result != nil) {
            test.Fatalf("Should not get a grading result.");
        }

        if (reject == nil) {
            test.Fatalf("Submission was not rejected when it should have been.");
        }

        if (!reflect.DeepEqual(expectedRejection, reject)) {
            test.Fatalf("Did not get the expected rejection. Expected: '%+v', Actual: '%+v'.", expectedRejection, reject);
        }
    }

    return result, reject, err;
}
