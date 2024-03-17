package common

import (
    "reflect"
    "testing"
    "time"

    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/util"
)

const ASSIGNMENT_ID = "exists";

func TestLogQueryBase(test *testing.T) {
    testCases := []struct{query RawLogQuery; expected ParsedLogQuery; errors []string; replaceTime bool}{
        {
            RawLogQuery{},
            ParsedLogQuery{},
            []string{},
            false,
        },

        {
            RawLogQuery{
                LevelString: log.LEVEL_STRING_INFO,
            },
            ParsedLogQuery{},
            []string{},
            false,
        },
        {
            RawLogQuery{
                LevelString: log.LEVEL_STRING_DEBUG,
            },
            ParsedLogQuery{
                Level: log.LevelDebug,
            },
            []string{},
            false,
        },
        {
            RawLogQuery{LevelString: "ZZZ"},
            ParsedLogQuery{},
            []string{
                "Could not parse 'level' component of log query ('ZZZ'): 'Unknown log level 'ZZZ'.'.",
            },
            false,
        },

        {
            RawLogQuery{
                AfterString: "2000-01-02T03:04:05Z",
            },
            ParsedLogQuery{
                After: MustTimestampFromString("2000-01-02T03:04:05Z").MustTime(),
            },
            []string{},
            false,
        },
        {
            RawLogQuery{
                AfterString: "2000-01-02",
            },
            ParsedLogQuery{},
            []string{
                `Could not parse 'after' component of log query ('2000-01-02'): 'Failed to parse timestamp string '2000-01-02': 'parsing time "2000-01-02" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"'.'.`,
            },
            false,
        },

        {
            RawLogQuery{
                PastString: "24h",
            },
            ParsedLogQuery{
                After: MustTimestampFromString("2000-01-02T03:04:05Z").MustTime(),
            },
            []string{},
            // Do not try to match the actual time (it will be flaky), just replace it.
            true,
        },
        {
            RawLogQuery{
                PastString: "-24h",
            },
            ParsedLogQuery{},
            []string{
                "Negative duration given for 'past' component of log query ('-24h').",
            },
            false,
        },
        {
            RawLogQuery{
                PastString: "ZZZ",
            },
            ParsedLogQuery{},
            []string{
                `Could not parse 'past' component of log query ('ZZZ'): 'time: invalid duration "ZZZ"'.`,
            },
            false,
        },

        {
            RawLogQuery{
                AfterString: "2000-01-02T03:04:05Z",
                PastString: "24h",
            },
            ParsedLogQuery{
                After: MustTimestampFromString("2000-01-02T03:04:05Z").MustTime(),
            },
            []string{},
            // Do not try to match the actual time (it will be flaky), just replace it.
            true,
        },

        {
            RawLogQuery{
                AssignmentID: ASSIGNMENT_ID,
            },
            ParsedLogQuery{
                AssignmentID: ASSIGNMENT_ID,
            },
            []string{},
            false,
        },
        {
            RawLogQuery{
                AssignmentID: "!!!",
            },
            ParsedLogQuery{},
            []string{
                "Could not parse 'assignment' component of log query ('!!!'): 'IDs must only have letters, digits, and single sequences of periods, underscores, and hyphens, found '!!!'.'.",
            },
            false,
        },
        {
            RawLogQuery{
                AssignmentID: "ZZZ",
            },
            ParsedLogQuery{},
            []string{
                "Unknown assignment given for 'assignment' component of log query ('ZZZ').",
            },
            false,
        },

        {
            RawLogQuery{
                TargetUser: "admin@test.com",
            },
            ParsedLogQuery{
                UserID: "admin@test.com",
            },
            []string{},
            false,
        },
    };

    for i, testCase := range testCases {
        actual, errors := testCase.query.ParseStrings(fakeCourse{});

        if (!reflect.DeepEqual(testCase.errors, errors)) {
            test.Fatalf("Case %d: Errors not as expected. Expected: '%s', Actual: '%s'.", i,
                    util.MustToJSONIndent(testCase.errors), util.MustToJSONIndent(errors));
        }

        if (testCase.replaceTime) {
            actual.After = testCase.expected.After;
        }

        if (testCase.expected != *actual) {
            test.Fatalf("Case %d: Parsed query not as expected. Expected: '%s', Actual: '%s'.", i,
                    util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(*actual));
        }
    }
}

func TestLogQueryParseTiming(test *testing.T) {
    now := MustTimestampFromString("2000-01-02T03:04:05Z").MustTime();

    testCases := []struct{afterString string; pastString string; expected time.Time}{
        {"", "", time.Time{}},
        {"2000-01-01T03:04:05Z", "", MustTimestampFromString("2000-01-01T03:04:05Z").MustTime()},
        {"", "24h", MustTimestampFromString("2000-01-01T03:04:05Z").MustTime()},
        {"2000-01-01T03:04:05Z", "1h", MustTimestampFromString("2000-01-02T02:04:05Z").MustTime()},
    };

    for i, testCase := range testCases {
        actual, err := ParseLogQueryTiming(now, testCase.afterString, testCase.pastString);
        if (err != nil) {
            test.Fatalf("Case %d: Unexpected error: '%v'.", i, err);
        }

        if (!testCase.expected.Equal(actual)) {
            test.Fatalf("Case %d: Time mismatch. Expected: '%s', Actual: '%s'.", i,
                    testCase.expected.String(), actual.String());
        }
    }
}

type fakeCourse struct{}

func (this fakeCourse) HasAssignment(id string) bool {
    return (id == ASSIGNMENT_ID);
}
