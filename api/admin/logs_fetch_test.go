package admin

import (
    "reflect"
    "slices"
    "testing"
    "time"

    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/util"
)

func TestFetchLogs(test *testing.T) {
    timeBeforeLogs := time.Now().Format(time.RFC3339);

    oldValue := log.SetBackgroundLogging(false);
    defer log.SetBackgroundLogging(oldValue);

    log.SetLevels(log.LevelOff, log.LevelTrace);
    defer log.SetLevelFatal();

    // Wait for old logs to get written.
    time.Sleep(10 * time.Millisecond)

    db.ResetForTesting();
    defer db.ResetForTesting();

    // Ignore logs with these messages.
    ignoreMessages := []string{
        "Loaded course.",
        "API Error",
    };

    course := db.MustGetTestCourse();

    log.Trace("trace", course);
    log.Debug("debug", course);
    log.Info("info", course);
    log.Warn("warn", course);
    log.Error("error", course);

    allRecords := []*log.Record{
        &log.Record{
            log.LevelTrace,
            "trace",
            0, "",
            course.GetID(), "", "",
            nil,
        },
        &log.Record{
            log.LevelDebug,
            "debug",
            0, "",
            course.GetID(), "", "",
            nil,
        },
        &log.Record{
            log.LevelInfo,
            "info",
            0, "",
            course.GetID(), "", "",
            nil,
        },
        &log.Record{
            log.LevelWarn,
            "warn",
            0, "",
            course.GetID(), "", "",
            nil,
        },
        &log.Record{
            log.LevelError,
            "error",
            0, "",
            course.GetID(), "", "",
            nil,
        },
    };

    timeAfterLogs := time.Now().Add(10 * time.Second).Format(time.RFC3339);

    testCases := []struct{
            role model.UserRole
            permError bool
            level string
            after string
            assignment string
            user string
            expectedErrors []string
            expectedRecords []*log.Record
    }{
        {model.RoleGrader, true, "", "", "", "", nil, nil},

        {model.RoleAdmin, false, "", "", "", "", nil, allRecords[2:]},
        {model.RoleAdmin, false, "trace", "", "", "", nil, allRecords},

        {model.RoleAdmin, false, "", timeBeforeLogs, "", "", nil, allRecords[2:]},
        {model.RoleAdmin, false, "", timeAfterLogs, "", "", nil, []*log.Record{}},

        // Parse Errors.
        {model.RoleAdmin, false, "ZZZ", "", "", "", []string{"Could not parse 'level': 'Unknown log level 'ZZZ'.'."}, nil},
        {model.RoleAdmin, false, "", "ZZZ", "", "", []string{"Could not parse 'after': 'ZZZ'."}, nil},
        {model.RoleAdmin, false, "", "", "!ZZZ", "", []string{"Improperly formatted 'assignment-id': 'IDs must only have letters, digits, and single sequences of periods, underscores, and hyphens, found '!zzz'.'."}, nil},
        {model.RoleAdmin, false, "", "", "ZZZ", "", []string{"Unknown assignment: 'zzz'."}, nil},
        {model.RoleAdmin, false, "", "", "", "ZZZ", []string{"Could not find user: 'ZZZ'."}, nil},
    };

    for i, testCase := range testCases {
        fields := map[string]any{
            "level": testCase.level,
            "after": testCase.after,
            "assignment-id": testCase.assignment,
            "target-email": testCase.user,
        };

        response := core.SendTestAPIRequestFull(test, core.NewEndpoint(`admin/logs/fetch`), fields, nil, testCase.role);
        if (!response.Success) {
            if (testCase.permError) {
                expectedLocator := "-020";
                if (response.Locator != expectedLocator) {
                    test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
                            i, expectedLocator, response.Locator);
                }
            } else {
                test.Errorf("Case %d: Response is not a success when it should be: '%v'.", i, response);
            }

            continue;
        }

        var responseContent FetchLogsResponse;
        util.MustJSONFromString(util.MustToJSON(response.Content), &responseContent);

        expectedSuccess := (len(testCase.expectedErrors) == 0);
        if (responseContent.Success != expectedSuccess) {
            test.Errorf("Case %d: Response success status is not as expected. Expected: '%v', Actual: '%v'.",
                    i, expectedSuccess, responseContent.Success);
            continue;
        }

        if (!reflect.DeepEqual(testCase.expectedErrors, responseContent.ErrorMessages)) {
            test.Errorf("Case %d: Error messages not as expected. Expected: '%v', Actual: '%v'.",
                    i, testCase.expectedErrors, responseContent.ErrorMessages);
            continue;
        }

        if (!responseContent.Success) {
            continue;
        }

        // Remove timestamps.
        for _, record := range responseContent.Records {
            record.UnixMicro = 0;
        }

        // Filter out records not related to this test.
        actualRecords := make([]*log.Record, 0, len(responseContent.Records));
        for _, record := range responseContent.Records {
            if (!slices.Contains(ignoreMessages, record.Message)) {
                actualRecords = append(actualRecords, record);
            }
        }

        if (!reflect.DeepEqual(testCase.expectedRecords, actualRecords)) {
            test.Errorf("Case %d: Log records not as expected. Expected: '%s', Actual: '%s'.",
                    i, util.MustToJSONIndent(testCase.expectedRecords), util.MustToJSONIndent(actualRecords));
            continue;
        }
    }
}
