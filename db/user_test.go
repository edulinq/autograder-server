package db

import (
    "reflect"
    "testing"

    "github.com/edulinq/autograder/email"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/util"
)

type SyncNewUsersTestCase struct {
    merge bool
    dryRun bool
    sendEmails bool
}

func (this *DBTests) DBTestResolveUsers(test *testing.T) {
    defer ResetForTesting();

    testCases := []struct {input []string; expectedOutput []string; addUsers []*model.User; removeUsers []string} {
        // This is a simple test case for the empty string input.
        {
            []string{""},
            []string{},
            nil,
            []string{},
        },

        // This is a test to ensure the output is sorted.
        {
            []string{"b@test.com", "a@test.com", "c@test.com"},
            []string{"a@test.com", "b@test.com", "c@test.com"},
            nil,
            []string{},
        },

        // This is a basic test to ensure that a role gets mapped to the correct email.
        {
            []string{"admin"},
            []string{"admin@test.com"},
            nil,
            []string{},
        },

        // This is a test for our all roles character, the *. 
        {
            []string{"*"},
            []string{"admin@test.com", "grader@test.com", "other@test.com", "owner@test.com", "student@test.com"},
            nil,
            []string{},
        },

        // This is a more complex test that adds multiple Users for the student role.
        // This test case removes the only user from the "other" role, so we check that a role without any users still functions properly.
        // This test case is given duplicate emails of different capitalizations and duplicate roles.
        // It tests to ensures we do not produce duplicates on this input.
        {
            []string{"other", "*", "grader@test.com", "zoinks@test.com", "ZoinKS@teSt.Com"},
            []string{"admin@test.com", "grader@test.com", "other@test.com", "second_student@test.com", "student@test.com", "zoinks@test.com"},
            []*model.User{model.NewUser("second_student@test.com", "", model.GetRole("student"))},
            []string{"owner@test.com"},
        },

        // This test case tests if miscapitalized roles still function and that we warn on invalid roles.
        {
            []string{"OTHER", "garbage"},
            []string{"other@test.com"},
            nil,
            []string{},
        },
    };

    for i, testCase := range testCases {
        ResetForTesting();
        course := MustGetCourse(TEST_COURSE_ID);

        for _, newUser := range testCase.addUsers {
            SaveUser(course, newUser);
        }

        for _, oldUser := range testCase.removeUsers {
            RemoveUser(course, oldUser);
        }

        actualOutput, err := ResolveUsers(course, testCase.input);
        if (err != nil) {
            test.Errorf("Case %d (%+v): Resolve User failed: '%v'.", i, testCase, err);
            continue;
        }

        if (!reflect.DeepEqual(testCase.expectedOutput, actualOutput)) {
            test.Errorf("Case %d (%+v): Incorrect Output. Expected: '%v', Actual: '%v'.", i,
                testCase, testCase.expectedOutput, actualOutput);
            continue;
        }
    }
}

func (this *DBTests) DBTestCourseSyncNewUsers(test *testing.T) {
    defer ResetForTesting();

    course := MustGetCourse(TEST_COURSE_ID);

    for i, testCase := range getSyncNewUsersTestCases() {
        ResetForTesting();

        testUsers, addUsers, shortCleartextPassUsers, fullCleartextPassUsers, shortEmails, fullEmails, modUsers, skipUsers := getSynNewUsersTestUsers();

        email.ClearTestMessages();

        result, err := SyncUsers(course, testUsers, testCase.merge, testCase.dryRun, testCase.sendEmails);
        if (err != nil) {
            test.Errorf("Case %d (%+v): User sync failed: '%v'.", i, testCase, err);
            continue;
        }

        testMessages := email.GetTestMessages();

        // New user syncs cannot delete users.
        if (len(result.Del) != 0) {
            test.Errorf("Case %d (%+v): Deleted users is not empty.", i, testCase);
            continue;
        }

        // There will always be adds.

        if (!model.UsersPointerEqual(addUsers, result.Add)) {
            test.Errorf("Case %d (%+v): Unexpected add users. Expected: '%s', actual: '%s'.",
                    i, testCase, util.MustToJSON(addUsers), util.MustToJSON(result.Add));
            continue;
        }

        // Depending on merge, a user will either be mod or skipped.
        // This also changed the emails sent.

        activeCase := result.Skip;
        activeName := "skip"
        activeExpected := skipUsers;
        emptyCase := result.Mod;
        emptyName := "mod";
        emails := shortEmails;
        cleartextPassUsers := shortCleartextPassUsers;

        if (testCase.merge) {
            activeCase = result.Mod;
            activeName = "mod"
            activeExpected = modUsers;
            emptyCase = result.Skip;
            emptyName = "skip";
            emails = fullEmails;
            cleartextPassUsers = fullCleartextPassUsers;
        }

        if (!model.UsersPointerEqual(activeExpected, activeCase)) {
            test.Errorf("Case %d (%+v): Unexpected %s users. Expected: '%s', actual: '%s'.",
                    i, testCase, activeName, util.MustToJSON(activeExpected), util.MustToJSON(activeCase));
            continue;
        }

        if (len(emptyCase) != 0) {
            test.Errorf("Case %d (%+v): Should have found 0 %s users, found (%d): '%s'.",
                    i, testCase, emptyName, len(emptyCase), util.MustToJSON(emptyCase));
            continue;
        }

        if (testCase.dryRun || !testCase.sendEmails) {
            if (len(testMessages) != 0) {
                test.Errorf("Case %d (%+v): Found %d emails when none should have been sent.", i, testCase, len(testMessages));
                continue;
            }
        } else {
            if (!email.ShallowSliceEqual(emails, testMessages)) {
                test.Errorf("Case %d (%+v): Unexpected emails. Expected: '%s', actual: '%s'.",
                        i, testCase, util.MustToJSON(emails), util.MustToJSON(testMessages));
                continue;
            }
        }

        if (len(cleartextPassUsers) != len(result.ClearTextPasswords)) {
            test.Errorf("Case %d (%+v): Number of cleartext passwords not as expected. Expected: %d, found: %d.",
                    i, testCase, len(cleartextPassUsers), len(result.ClearTextPasswords));
            continue;
        }

        for _, email := range cleartextPassUsers {
            _, ok := result.ClearTextPasswords[email];
            if (!ok) {
                test.Errorf("Case %d (%+v): User '%s' does not have a cleartext password.", i, testCase, email);
                continue;
            }
        }
    }
}

// Get all possible test cases.
func getSyncNewUsersTestCases() []SyncNewUsersTestCase {
    return buildSyncNewUsersTestCase(nil, 0, make([]bool, 5));
}

func buildSyncNewUsersTestCase(testCases []SyncNewUsersTestCase, index int, currentCase []bool) []SyncNewUsersTestCase {
    if (index >= len(currentCase)) {
        return append(testCases, SyncNewUsersTestCase{
            merge: currentCase[0],
            dryRun: currentCase[1],
            sendEmails: currentCase[2],
        });
    }

    currentCase[index] = true;
    testCases = buildSyncNewUsersTestCase(testCases, index + 1, currentCase);

    currentCase[index] = false;
    testCases = buildSyncNewUsersTestCase(testCases, index + 1, currentCase);

    return testCases;
}

func getSynNewUsersTestUsers() (
        map[string]*model.User, []*model.User, []string, []string, []*email.Message, []*email.Message, []*model.User, []*model.User) {
    var testUsers map[string]*model.User = map[string]*model.User{
        "add@test.com": &model.User{
            Email: "add@test.com",
            Name: "add",
            LMSID: "lms-add@test.com",
        },
        "add-pass@test.com": &model.User{
            Email: "add-pass@test.com",
            Pass: util.Sha256HexFromString("add-pass"),
            Name: "add pass",
            Role: model.RoleStudent,
            LMSID: "lms-add-pass@test.com",
        },
        "other@test.com": &model.User{
            Email: "other@test.com",
            Name: "modified",
            Role: model.RoleStudent,
            LMSID: "lms-mod@test.com",
        },
        "student@test.com": &model.User{
            Email: "student@test.com",
            Pass: util.Sha256HexFromString("mod-pass"),
        },
        // No change, should be marked as mod (because of password).
        "grader@test.com": &model.User{
            Email: "grader@test.com",
            // No role change.
            Role: model.RoleUnknown,
        },
    };

    // The users that are marked as additions.
    var addUsers []*model.User = []*model.User{
        &model.User{
            Email: "add@test.com",
            Name: "add",
            Role: model.RoleOther,
            LMSID: "lms-add@test.com",
        },
        &model.User{
            Email: "add-pass@test.com",
            Name: "add pass",
            Role: model.RoleStudent,
            LMSID: "lms-add-pass@test.com",
        },
    };

    // The users that will have cleartext passwords when users are skipped.
    var shortCleartextPassUsers []string = []string{"add@test.com"};

    // The users that will have cleartext passwords when users are merged.
    var fullCleartextPassUsers []string = []string{"add@test.com", "other@test.com", "grader@test.com"};

    // The emails when users are skipped.
    var shortEmails []*email.Message = []*email.Message{
        &email.Message{
            To: []string{"add@test.com"},
            Subject: "Autograder course101 -- User Account Created",
            HTML: false,
        },
        &email.Message{
            To: []string{"add-pass@test.com"},
            Subject: "Autograder course101 -- User Account Created",
            HTML: false,
        },
    };

    // The emails when users are merged.
    var fullEmails []*email.Message = []*email.Message{
        &email.Message{
            To: []string{"add@test.com"},
            Subject: "Autograder course101 -- User Account Created",
            HTML: false,
        },
        &email.Message{
            To: []string{"add-pass@test.com"},
            Subject: "Autograder course101 -- User Account Created",
            HTML: false,
        },
        &email.Message{
            To: []string{"other@test.com"},
            Subject: "Autograder course101 -- User Password Changed",
            HTML: false,
        },
        &email.Message{
            To: []string{"grader@test.com"},
            Subject: "Autograder course101 -- User Password Changed",
            HTML: false,
        },
    };

    // The users that are marked as mods.
    // These will not appear in every case.
    var modUsers []*model.User = []*model.User{
        &model.User{
            Email: "other@test.com",
            Name: "modified",
            Role: model.RoleStudent,
            LMSID: "lms-mod@test.com",
        },
        &model.User{
            Email: "student@test.com",
            Name: "student",
            Role: model.RoleStudent,
            LMSID: "",
        },
        &model.User{
            Email: "grader@test.com",
            Name: "grader",
            Role: model.RoleGrader,
            LMSID: "",
        },
    };

    // The users that are marked as skips.
    // These will not appear in every case.
    var skipUsers []*model.User = []*model.User{
        &model.User{
            Email: "other@test.com",
            Name: "other",
            Role: model.RoleOther,
        },
        &model.User{
            Email: "student@test.com",
            Name: "student",
            Role: model.RoleStudent,
        },
        &model.User{
            Email: "grader@test.com",
            Name: "grader",
            Role: model.RoleGrader,
        },
    };

    return testUsers, addUsers, shortCleartextPassUsers, fullCleartextPassUsers, shortEmails, fullEmails, modUsers, skipUsers;
}
