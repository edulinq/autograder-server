package db

import (
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/util"
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

// Tests GetSubmissionAttempts as follows:
// A) Fetch all attempts from a user who has submissions and check that the result is not empty.
// B) Fetch attempts from a user who has no submissions and check that the result is empty.
// C) Make a submission to a user with no entries, then fetch that attempt and makes sure the result has one entry.
// D) Remove that submission and then fetch again and maker sure the result is empty.
func (this *DBTests) DBTestFetchAttempts(test *testing.T) {
    ResetForTesting();
    defer ResetForTesting();

    assignment := MustGetTestAssignment();

    // Case A
    studentAttempts, err := GetSubmissionAttempts(assignment, "student@test.com");
    if err != nil {
        test.Fatalf("Failed to get student attempts when should be a success: '%v'.", err);
    }

    if len(studentAttempts) == 0 {
        test.Fatalf("Got an empty result when shouldn't be.");
    }

    // Case B
    graderAttempts, err := GetSubmissionAttempts(assignment, "grader@test.com");
    if err != nil {
        test.Fatalf("Failed to get grader attempts when should be a success (with empty result): '%v'.", err);
    }

    if len(graderAttempts) != 0 {
        test.Fatalf("Unexpected result length. Expected: '%d', Actual: '%d'.", 0, len(graderAttempts));
    }

    // Case C
    graderSubmission := studentAttempts[0];
    graderSubmission.Info.User = "grader@test.com";

    err = SaveSubmission(assignment, graderSubmission);
    if err != nil {
        test.Fatalf("Failed to save grader submission: '%v'.", err);
    }

    graderAttempts, err = GetSubmissionAttempts(assignment, "grader@test.com");
    if err != nil {
        test.Fatalf("Failed to get grader attempts when there should be one: '%v'.", err);
    }

    if len(graderAttempts) != 1 {
        test.Fatalf("Fetch returned unexpected number of attempts. Expected: '%d', Actual: '%d'.", 1, len(graderAttempts));
    }

    if (!reflect.DeepEqual(graderAttempts[0], graderSubmission)) {
        test.Errorf("Unexpected attempt returned. Expected: '%s', actual: '%s'.",
            util.MustToJSONIndent(graderAttempts[0]), util.MustToJSONIndent(graderSubmission));
    }

    // Case D
    isRemoved, err := RemoveSubmission(assignment, "grader@test.com", "");
    if err != nil {
        test.Fatalf("Failed to remove grader submission: '%v'.", err);
    }

    if !isRemoved {
        test.Fatalf("Returned false from RemoveSubmission() when should be true.");
    }

    graderAttempts, err = GetSubmissionAttempts(assignment, "grader@test.com");
    if err != nil {
        test.Fatalf("Failed to get grader attempts when should be a success (with empty result): '%v'.", err);
    }

    if len(graderAttempts) != 0 {
        test.Fatalf("Unexpected result length. Expected: '%d', Actual: '%d'.", 0, len(graderAttempts));
    }
}
