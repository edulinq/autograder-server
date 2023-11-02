package lms

import (
    "reflect"
    "testing"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/grader"
    lmstest "github.com/eriq-augustine/autograder/lms/adapter/test"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

func TestUploadScores(test *testing.T) {
    course := grader.GetCourse("course-with-lms");
    if (course == nil) {
        test.Fatalf("Failed to get test course.");
    }

    testLMSAdapter := course.GetLMSAdapter().Adapter.(*lmstest.TestLMSAdapter);

    // Reset the LMS adapter.
    defer func() {
        testLMSAdapter.FailUpdateAssignmentScores = false;
    }();

    testCases := []struct{ role usr.UserRole; permError bool; failUpdate bool; scores []ScoreEntry; expected *UploadScoresResponse }{
        // Normal.
        {
            usr.Grader, false, false,
            []ScoreEntry{
                ScoreEntry{"student@test.com", 10},
            },
            &UploadScoresResponse{
                Count: 1,
                ErrorCount: 0,
                UnrecognizedUsers: nil,
                NoLMSIDUsers: nil,
            },
        },
        {
            usr.Grader, false, false,
            []ScoreEntry{
                ScoreEntry{"student@test.com", 10},
                ScoreEntry{"grader@test.com", 0},
                ScoreEntry{"admin@test.com", -10},
                ScoreEntry{"owner@test.com", 12.34},
            },
            &UploadScoresResponse{
                Count: 4,
                ErrorCount: 0,
                UnrecognizedUsers: nil,
                NoLMSIDUsers: nil,
            },
        },

        // Permissions.
        {usr.Other, true, false, nil, nil},
        {usr.Student, true, false, nil, nil},

        // Upload fails.
        {
            usr.Grader, false, true,
            []ScoreEntry{
                ScoreEntry{"student@test.com", 10},
            },
            nil,
        },

        // Bad scores.
        {
            usr.Grader, false, false,
            []ScoreEntry{
                ScoreEntry{"zzz@test.com", 10},
                ScoreEntry{"no-lms-id@test.com", 20},
                ScoreEntry{"abc@test.com", 30},
                ScoreEntry{"student@test.com", 10},
            },
            &UploadScoresResponse{
                Count: 1,
                ErrorCount: 3,
                UnrecognizedUsers: []RowEntry{
                    RowEntry{0, "zzz@test.com"},
                    RowEntry{2, "abc@test.com"},
                },
                NoLMSIDUsers: []RowEntry{
                    RowEntry{1, "no-lms-id@test.com"},
                },
            },
        },

        // Upload will pass, but never gets called.
        {
            usr.Grader, false, false,
            []ScoreEntry{
                ScoreEntry{"zzz@test.com", 10},
            },
            &UploadScoresResponse{
                Count: 0,
                ErrorCount: 1,
                UnrecognizedUsers: []RowEntry{
                    RowEntry{0, "zzz@test.com"},
                },
                NoLMSIDUsers: nil,
            },
        },

        // Upload will fail, but never gets called.
        {
            usr.Grader, false, true,
            []ScoreEntry{
                ScoreEntry{"zzz@test.com", 10},
            },
            &UploadScoresResponse{
                Count: 0,
                ErrorCount: 1,
                UnrecognizedUsers: []RowEntry{
                    RowEntry{0, "zzz@test.com"},
                },
                NoLMSIDUsers: nil,
            },
        },
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "course-id": "course-with-lms",
            // ID does not matter, test LMS will accept all ids.
            "assignment-lms-id": "foo",
            "scores": testCase.scores,
        };

        testLMSAdapter.FailUpdateAssignmentScores = testCase.failUpdate;

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`lms/upload/scores`), fields, nil, testCase.role);
        if (!response.Success) {
            expectedLocator := "";
            if (testCase.permError) {
                expectedLocator = "-306";
            } else if (testCase.failUpdate) {
                expectedLocator = "-506";
            }

            if (expectedLocator == "") {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            } else {
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            }

            continue;
        }

        var responseContent UploadScoresResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        if (!reflect.DeepEqual(testCase.expected, &responseContent)) {
            test.Errorf("Case %d: Unexpected result. Expected: '%+v', actual: '%+v'.", i, testCase.expected, responseContent);
            continue;
        }
    }
}
