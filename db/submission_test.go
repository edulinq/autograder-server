package db

import (
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
