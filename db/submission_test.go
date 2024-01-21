package db

import (
    "github.com/eriq-augustine/autograder/util"
    "reflect"
    "testing"
)

func (this *DBTests) DBTestRemoveSubmission(test *testing.T) {
    defer ResetForTesting();

    testCases := []struct{targetEmail string; targetSubmission string; isRemoved bool} {
        // Specific email, specific submission.
        {"student@test.com", "1697406256", true},
        {"student@test.com", "1697406265", true},
        {"student@test.com", "1697406272", true},

        // Specific email, specific submission (full ID).
        {"student@test.com", "course101::hw0::student@test.com::1697406256", true},
        {"student@test.com", "course101::hw0::student@test.com::1697406265", true},
        {"student@test.com", "course101::hw0::student@test.com::1697406272", true},

        // Specific email, missing submission.
        {"student@test.com", "ZZZ", false},

        // Specific email, recent submission.
        {"student@test.com", "", true},

        // Missing email, specific submission.
        {"ZZZ@test.com", "1697406256", false},
        {"ZZZ@test.com", "1697406265", false},
        {"ZZZ@test.com", "1697406272", false},

        // Missing email, missing submission.
        {"ZZZ@test.com", "ZZZ", false},

        // Missing email, specific submission (full ID).
        {"ZZZ@test.com", "course101::hw0::student@test.com::1697406256", false},
        {"ZZZ@test.com", "course101::hw0::student@test.com::1697406265", false},
        {"ZZZ@test.com", "course101::hw0::student@test.com::1697406272", false},

        // Missing email, recent submission.
        {"ZZZ@test.com", "", false},
    };

    for i, testCase := range testCases {
        // Reload the test course every time.
        ResetForTesting();
        
        assignment := MustGetTestAssignment();

        isRemoved, err := RemoveSubmission(assignment, testCase.targetEmail, testCase.targetSubmission);
        if (err != nil) {
            test.Errorf("Case %d: Submission removal failed: '%v'.", i, err);
            continue;
        }

        if (isRemoved != testCase.isRemoved) {
            test.Errorf("Case %d: Removed submission does not match. Expected : '%v', actual: '%v'.", i, testCase.isRemoved, isRemoved);
            continue;
        }
    }
}

func (this *DBTests) DBTestFetchAttempts(test *testing.T) {
    ResetForTesting();
    defer ResetForTesting();

    assignment := MustGetTestAssignment();

    // Fetch attempts from another user who has submissions.
    studentAttempts, err := GetSubmissionAttempts(assignment, "student@test.com");
    if err != nil {
        test.Fatalf("Failed to get student attempts when should be a success: %v", err);
    }

    if len(studentAttempts) == 0 {
        test.Fatalf("Got an empty result when shouldn't be.");
    }

    // Fetch attempts from self (grader). User does not have submissions.
    graderAttempts, err := GetSubmissionAttempts(assignment, "grader@test.com");
    if err != nil {
        test.Fatalf("Failed to get grader attempts when should be a success (with empty result): %v", err);
    }

    if len(graderAttempts) != 0 {
        test.Fatalf("Got a non-empty result when should be empty.");
    }

    // User makes a submission, then that submission is removed.
    // This case tests GetSubmissionAttempts() when the submission directory exists but is empty.
    graderSubmission := studentAttempts[0];
    graderSubmission.Info.User = "grader@test.com";

    err = SaveSubmission(assignment, graderSubmission);
    if err != nil {
        test.Fatalf("Failed to save grader submission %v: ", err);
    }

    graderAttempts, err = GetSubmissionAttempts(assignment, "grader@test.com");
    if err != nil {
        test.Fatalf("Failed to get grader attempts when there should be one %v: ", err);
    }

    if len(graderAttempts) != 1 {
        test.Fatalf("Fetch returned unexpected number of attempts. Expected: %d, actual: %d", 1, len(graderAttempts));
    }

    if (!reflect.DeepEqual(graderAttempts[0], graderSubmission)) {
        test.Errorf("Unexpected attempt returned. Expected: '%s', actual: '%s'.",
            util.MustToJSONIndent(graderAttempts[0]), util.MustToJSONIndent(graderSubmission));
    }

    isRemoved, err := RemoveSubmission(assignment, "grader@test.com", "");
    if err != nil {
        test.Fatalf("Failed to remove grader submission: %v", err);
    }

    if !isRemoved {
        test.Fatalf("Returned false from RemoveSubmission() when should be true.");
    }

    graderAttempts, err = GetSubmissionAttempts(assignment, "grader@test.com");
    if err != nil {
        test.Fatalf("Failed to get grader attempts when should be a success (with empty result): %v", err);
    }

    if len(graderAttempts) != 0 {
        test.Fatalf("Got a non-empty result when should be.");
    }
}
