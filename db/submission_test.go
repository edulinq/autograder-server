package db

import (
    "testing"
)

func (this *DBTests) DBTestRemoveSubmission(test *testing.T) {
    defer ResetForTesting();

    testCases := []struct{targetEmail string; targetSubmission string; isRemoved bool}{

        // Student, specific.
        {"student@test.com", "1697406256", true},
        {"student@test.com", "1697406265", true},
        {"student@test.com", "1697406272", true},

        // Student, specific (full ID).
        {"student@test.com", "course101::hw0::student@test.com::1697406256", true},
        {"student@test.com", "course101::hw0::student@test.com::1697406265", true},
        {"student@test.com", "course101::hw0::student@test.com::1697406272", true},

        // Student, missing.
        {"student@test.com", "ZZZ", false},

        // Student, recent.
        {"student@test.com", "", true},

        // Missing, recent.
        {"ZZZ@test.com", "", false},

        // Missing, missing.
        {"ZZZ@test.com", "", false},

        // Missing, specific.
        {"ZZZ@test.com", "1697406256", false},
        {"ZZZ@test.com", "1697406265", false},
        {"ZZZ@test.com", "1697406272", false},
    };

    for i, testCase := range testCases {
        // Reload the test course every time.
        ResetForTesting();
        
        assignment := MustGetAssignment(TEST_COURSE_ID, "hw0");

        isRemoved, err := RemoveSubmission(assignment, testCase.targetEmail, testCase.targetSubmission);

        if (err != nil) {
            test.Errorf("Case %d: Submission removal failed: '%v", i, err);
            continue;
        }

        if (isRemoved != testCase.isRemoved) {
            test.Errorf("Case %d: Removed submission does not match. Expected : '%v', actual: '%v'.", i, testCase.isRemoved, isRemoved);
            continue;
        }
    }
}
