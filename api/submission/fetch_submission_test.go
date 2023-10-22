package submission

import (
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

var studentHashes map[string]string = map[string]string{
    "1697406256": "b1bfd6d119e235ba086b1403b3c56b126e2317fea27ab14dbd73178be1057eea",
    "1697406265": "9215f28f8c88d15a2edee9cdbad20ccab34b187cd9d77823a4925d46971f755b",
    "1697406272": "1b25084ce5a6e85a5e7e9660e295f19769db08335d544ba6b900185498c4944a",
}

func TestFetchSubmission(test *testing.T) {
    // See writeZipContents() for hash information.
    testCases := []struct{
            role usr.UserRole
            targetEmail string
            targetSubmission string
            foundUser bool
            foundSubmission bool
            permError bool
            hash string
    }{
        // Grader, self, recent.
        {usr.Grader, "",                "", true, false, false, ""},
        {usr.Grader, "grader@test.com", "", true, false, false, ""},

        // Grader, self, missing.
        {usr.Grader, "",                "ZZZ", true, false, false, ""},
        {usr.Grader, "grader@test.com", "ZZZ", true, false, false, ""},

        // Grader, other, recent.
        {usr.Grader, "student@test.com", "", true, true, false, studentHashes["1697406272"]},

        // Grader, other, specific.
        {usr.Grader, "student@test.com", "1697406256", true, true, false, studentHashes["1697406256"]},
        {usr.Grader, "student@test.com", "1697406265", true, true, false, studentHashes["1697406265"]},
        {usr.Grader, "student@test.com", "1697406272", true, true, false, studentHashes["1697406272"]},

        // Grader, other, specific (full ID).
        {usr.Grader, "student@test.com", "course101::hw0::student@test.com::1697406256", true, true, false, studentHashes["1697406256"]},
        {usr.Grader, "student@test.com", "course101::hw0::student@test.com::1697406265", true, true, false, studentHashes["1697406265"]},
        {usr.Grader, "student@test.com", "course101::hw0::student@test.com::1697406272", true, true, false, studentHashes["1697406272"]},

        // Grader, other, missing.
        {usr.Grader, "student@test.com", "ZZZ", true, false, false, ""},

        // Grader, missing, recent.
        {usr.Grader, "ZZZ@test.com", "", false, false, false, ""},

        // Student, self, recent.
        {usr.Student, "",                 "", true, true, false, studentHashes["1697406272"]},
        {usr.Student, "student@test.com", "", true, true, false, studentHashes["1697406272"]},

        // Student, self, missing.
        {usr.Student, "",                 "ZZZ", true, false, false, ""},
        {usr.Student, "student@test.com", "ZZZ", true, false, false, ""},

        // Student, other, recent.
        {usr.Student, "grader@test.com", "", false, false, true, ""},

        // Student, other, missing.
        {usr.Student, "grader@test.com", "ZZZ", true, false, true, ""},
    };

    // Quiet the output a bit.
    oldLevel := config.GetLoggingLevel();
    config.SetLogLevelFatal();
    defer config.SetLoggingLevel(oldLevel);

    for i, testCase := range testCases {
        fields := map[string]any{
            "target-email": testCase.targetEmail,
            "target-submission": testCase.targetSubmission,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`submission/fetch/submission`), fields, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-319";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }

            continue;
        }

        var responseContent FetchSubmissionResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (testCase.foundUser != responseContent.FoundUser) {
            test.Errorf("Case %d: Found user does not match. Expected: '%v', actual: '%v'.", i, testCase.foundUser, responseContent.FoundUser);
            continue;
        }

        if (testCase.foundSubmission != responseContent.FoundSubmission) {
            test.Errorf("Case %d: Found submission does not match. Expected: '%v', actual: '%v'.", i, testCase.foundSubmission, responseContent.FoundSubmission);
            continue;
        }

        if (!responseContent.FoundSubmission) {
            continue;
        }

        actualHash := util.Sha256HexFromString(responseContent.Contents);
        if (testCase.hash != actualHash) {
            test.Errorf("Case %d: Unexpected submission hash. Expected: '%+v', actual: '%+v'.", i, testCase.hash, actualHash);
            continue;
        }
    }
}

// If you need to validate a new hash,
// then print the hash, write out the zip, and manually inspect the zip contents.
func writeZipContents(test *testing.T, responseContent *FetchSubmissionResponse) {
    data, err := util.Base64Decode(responseContent.Contents);
    if (err != nil) {
        test.Fatalf("Failed to decode: '%v'.", err);
    }

    err = util.WriteBinaryFile(data, "/tmp/test.zip");
    if (err != nil) {
        test.Fatalf("Failed to write: '%v'.", err);
    }
}
