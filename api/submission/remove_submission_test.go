package submission

import (
    "testing"
    
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
)

func TestRemoveSubmission(test *testing.T) {
	// Leave the course in a good state after the test.
	defer db.ResetForTesting();

    testCases := []struct{ role model.UserRole; targetEmail string; targetSubmission string; foundUser bool; foundSubmission bool; permError bool}{
        // Grader, self, recent.
        {model.RoleGrader, "grader@test.com", "", true, false, false},
        {model.RoleGrader, "grader@test.com", "", true, false, false},

        // Grader, self, missing.
        {model.RoleGrader, "",                "ZZZ", true, false, false},
        {model.RoleGrader, "grader@test.com", "ZZZ", true, false, false},

        // Grader, other, recent.
        {model.RoleGrader, "student@test.com", "",true, true, false},

        // Grader, other, specific.
        {model.RoleGrader, "student@test.com", "1697406256", true, true, false, },
        {model.RoleGrader, "student@test.com", "1697406265", true, true, false},
        {model.RoleGrader, "student@test.com", "1697406272", true, true, false},

        // Grader, other, specific (full ID).
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406256", true, true, false},
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406265", true, true, false},
        {model.RoleGrader, "student@test.com", "course101::hw0::student@test.com::1697406272", true, true, false},

        // Grader, other, missing.
        {model.RoleGrader, "student@test.com", "ZZZ", true, false, true},

        // Grader, missing, recent.
        {model.RoleGrader, "ZZZ@test.com", "", false, false, true},
    };

    for i, testCase := range testCases {
		// Reload the test course every time.
		db.ResetForTesting();

        fields := map[string]any{
            "target-email": testCase.targetEmail,
            "target-submission": testCase.targetSubmission,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/remove/submission`), fields, nil, testCase.role);
        
		if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-320";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }
			
            continue;
        }

    }

}
