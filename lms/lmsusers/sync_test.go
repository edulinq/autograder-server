package lmsusers

import (
    "slices"
    "testing"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/lms/lmstypes"
    lmstest "github.com/eriq-augustine/autograder/lms/backend/test"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

type SyncLMSTestCase struct {
    dryRun bool
    sendEmails bool
    syncAttributes bool
    syncAdd bool
    syncDel bool
}

func reset() {
    db.ResetForTesting();
    lmstest.ClearUsersModifier();
}

func TestCourseSyncLMSUisers(test *testing.T) {
    // Leave the db in a good state after the test.
    defer reset();

    for i, testCase := range getSyncLMSTestCases() {
        // Reload the test course every time.
        reset();

        lmstest.SetUsersModifier(testingUsers);
        course := db.MustGetTestCourse();

        course.GetLMSAdapter().SyncUserAttributes = testCase.syncAttributes;
        course.GetLMSAdapter().SyncAddUsers = testCase.syncAdd;
        course.GetLMSAdapter().SyncRemoveUsers = testCase.syncDel;

        localUsers := db.MustGetUsers(course);

        email.ClearTestMessages();

        result, err := SyncLMSUsers(course, testCase.dryRun, testCase.sendEmails);
        if (err != nil) {
            test.Errorf("Case %d (%+v): User sync failed: '%v'.", i, testCase, err);
            continue;
        }

        var unchangedUsers []*model.User = []*model.User{
            localUsers["owner@test.com"],
        };

        testMessages := email.GetTestMessages();

        // LMS syncs cannot skip users.
        if (len(result.Skip) != 0) {
            test.Errorf("Case %d (%+v): Skipped users is not empty.", i, testCase);
            continue;
        }

        // There will always be mod users, since LMS IDs are always synced.
        // But when the option is on, additional attriutes will be synced.
        currentModUsers := modUsers;
        if (testCase.syncAttributes) {
            currentModUsers = modAllUsers;
        } else {
        }

        if (!model.UsersPointerEqual(currentModUsers, result.Mod)) {
            test.Errorf("Case %d (%+v): Unexpected mod users. Expected: '%s', actual: '%s'.",
                    i, testCase, util.MustToJSON(currentModUsers), util.MustToJSON(result.Mod));
            continue;
        }

        if (testCase.syncAdd) {
            if (!model.UsersPointerEqual(addUsers, result.Add)) {
                test.Errorf("Case %d (%+v): Unexpected add users. Expected: '%s', actual: '%s'.",
                        i, testCase, util.MustToJSON(addUsers), util.MustToJSON(result.Add));
                continue;
            }

            if (len(result.Add) != len(result.ClearTextPasswords)) {
                test.Errorf("Case %d (%+v): Number of cleartext passwords (%d) does not match number of add users (%d).",
                        i, testCase, len(result.ClearTextPasswords), len(result.Add));
                continue;
            }

            for _, user := range addUsers {
                _, ok := result.ClearTextPasswords[user.Email];
                if (!ok) {
                    test.Errorf("Case %d (%+v): Add user '%s' does not have a cleartext password.", i, testCase, user.Email);
                    continue;
                }
            }

            if (testCase.dryRun || !testCase.sendEmails) {
                if (len(testMessages) != 0) {
                    test.Errorf("Case %d (%+v): User additions were enabled on a no-email/dry run, but %d new emails were found.", i, testCase, len(testMessages));
                    continue;
                }
            } else {
                if (!email.ShallowSliceEqual(addEmails, testMessages)) {
                    test.Errorf("Case %d (%+v): Unexpected add emails. Expected: '%s', actual: '%s'.",
                            i, testCase, util.MustToJSON(addEmails), util.MustToJSON(testMessages));
                    continue;
                }
            }
        } else {
            if (len(result.Add) != 0) {
                test.Errorf("Case %d (%+v): User additions were disabled, but %d new users were found.", i, testCase, len(result.Add));
                continue;
            }

            if (len(result.ClearTextPasswords) != 0) {
                test.Errorf("Case %d (%+v): User additions were disabled, but %d new cleartext passwords were found.", i, testCase, len(result.ClearTextPasswords));
                continue;
            }

            if (len(testMessages) != 0) {
                test.Errorf("Case %d (%+v): User additions were disabled, but %d new emails were found.", i, testCase, len(testMessages));
                continue;
            }
        }

        if (testCase.syncDel) {
            if (!model.UsersPointerEqual(delUsers, result.Del)) {
                test.Errorf("Case %d (%+v): Unexpected del users. Expected: '%s', actual: '%s'.",
                        i, testCase, util.MustToJSON(delUsers), util.MustToJSON(result.Del));
                continue;
            }
        } else {
            unchangedUsers = append(unchangedUsers, localUsers["other@test.com"]);

            if (len(result.Del) != 0) {
                test.Errorf("Case %d (%+v): User deletions were disabled, but %d deleted users were found.", i, testCase, len(result.Del));
                continue;
            }
        }

        if (!model.UsersPointerEqual(unchangedUsers, result.Unchanged)) {
            test.Errorf("Case %d (%+v): Unexpected unchanged users. Expected: '%s', actual: '%s'.",
                    i, testCase, util.MustToJSON(unchangedUsers), util.MustToJSON(result.Unchanged));
            continue;
        }
    }
}

// Get all possible test cases.
func getSyncLMSTestCases() []SyncLMSTestCase {
    return buildSyncLMSTestCase(nil, 0, make([]bool, 5));
}

func buildSyncLMSTestCase(testCases []SyncLMSTestCase, index int, currentCase []bool) []SyncLMSTestCase {
    if (index >= len(currentCase)) {
        return append(testCases, SyncLMSTestCase{
            dryRun: currentCase[0],
            sendEmails: currentCase[1],
            syncAttributes: currentCase[2],
            syncAdd: currentCase[3],
            syncDel: currentCase[4],
        });
    }

    currentCase[index] = true;
    testCases = buildSyncLMSTestCase(testCases, index + 1, currentCase);

    currentCase[index] = false;
    testCases = buildSyncLMSTestCase(testCases, index + 1, currentCase);

    return testCases;
}

// Modify the users that the LMS will return for testing.
func testingUsers(users []*lmstypes.User) []*lmstypes.User {
    // Remove other.
    removeIndex := -1;
    for i, user := range users {
        if (user.Email == "other@test.com") {
            removeIndex = i;
        } else if (user.Email == "student@test.com") {
            // student will only have their LMS ID added, no other changes.
        } else if (user.Email == "grader@test.com") {
            // grader will have their name changes.
            user.Name = "Changed Name";
        } else if (user.Email == "admin@test.com") {
            // admin will have their role changed.
            user.Role = model.RoleOwner;
        } else if (user.Email == "owner@test.com") {
            // owner will not have anything changed (so we must manually remove their LMS ID).
            user.ID = "";
        }
    }

    users = slices.Delete(users, removeIndex, removeIndex + 1);

    // Make an add user.
    addUser := &lmstypes.User{
        ID: "lms-add@test.com",
        Name: "add",
        Email: "add@test.com",
        Role: model.RoleStudent,
    };
    users = append(users, addUser);

    return users;
}

// The users that are marked as additions.
var addUsers []*model.User = []*model.User{
    &model.User{
        Email: "add@test.com",
        DisplayName: "add",
        Role: model.RoleStudent,
        LMSID: "lms-add@test.com",
    },
};

// The users that are marked as deletions.
var delUsers []*model.User = []*model.User{
    &model.User{
        Email: "other@test.com",
        DisplayName: "other",
        Role: model.RoleOther,
        LMSID: "",
    },
};

// All the users that are marked as mods.
var modAllUsers []*model.User = []*model.User{
    &model.User{
        Email: "student@test.com",
        DisplayName: "student",
        Role: model.RoleStudent,
        LMSID: "lms-student@test.com",
    },
    &model.User{
        Email: "grader@test.com",
        DisplayName: "Changed Name",
        Role: model.RoleGrader,
        LMSID: "lms-grader@test.com",
    },
    &model.User{
        Email: "admin@test.com",
        DisplayName: "admin",
        Role: model.RoleOwner,
        LMSID: "lms-admin@test.com",
    },
};

// All the users that are marked as mods with no attribute syncing.
var modUsers []*model.User = []*model.User{
    &model.User{
        Email: "student@test.com",
        DisplayName: "student",
        Role: model.RoleStudent,
        LMSID: "lms-student@test.com",
    },
    &model.User{
        Email: "grader@test.com",
        DisplayName: "grader",
        Role: model.RoleGrader,
        LMSID: "lms-grader@test.com",
    },
    &model.User{
        Email: "admin@test.com",
        DisplayName: "admin",
        Role: model.RoleAdmin,
        LMSID: "lms-admin@test.com",
    },
};

var addEmails []*email.Message = []*email.Message{
    &email.Message{
        To: []string{"add@test.com"},
        Subject: "Autograder -- User Account Created",
        HTML: false,
    },
};
