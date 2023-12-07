package tasks

import (
    "fmt"
    "testing"
    "time"
)

// 2023-10-01 00:00 Sunday
var baseTime time.Time = time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)

type timeSpecTestCase struct {
    ID string
    TimeSpec timeSpec
    NextTime time.Time
    TotalSecs int
    IsEmpty bool
    String string
}

func getValidTestCases() []*timeSpecTestCase {
    var testCases []*timeSpecTestCase;

    for i, testCase := range validDurationCases {
        testCases = append(testCases, &timeSpecTestCase{
            ID: fmt.Sprintf("Duration, Index %d", i),
            TimeSpec: testCase.TimeSpec,
            NextTime: testCase.NextTime,
            TotalSecs: testCase.TotalSecs,
            IsEmpty: testCase.IsEmpty,
            String: testCase.String,
        });

        if (testCase.TimeSpec.IsEmpty()) {
            continue;
        }

        testCases = append(testCases, &timeSpecTestCase{
            ID: fmt.Sprintf("ScheduledTime(Duration), Index %d", i),
            TimeSpec: &ScheduledTime{Every: *testCase.TimeSpec},
            NextTime: testCase.NextTime,
            TotalSecs: testCase.TotalSecs,
            IsEmpty: testCase.IsEmpty,
            String: testCase.String,
        });
    }

    for i, testCase := range validTimeOfDayCases {
        testCases = append(testCases, &timeSpecTestCase{
            ID: fmt.Sprintf("TimeOfDay, Index %d", i),
            TimeSpec: testCase.TimeSpec,
            NextTime: testCase.NextTime,
            TotalSecs: testCase.TotalSecs,
            IsEmpty: testCase.IsEmpty,
            String: testCase.String,
        });

        if (testCase.TimeSpec.IsEmpty()) {
            continue;
        }

        testCases = append(testCases, &timeSpecTestCase{
            ID: fmt.Sprintf("ScheduledTime(TimeOfDay), Index %d", i),
            TimeSpec: &ScheduledTime{Daily: testCase.TimeSpec},
            NextTime: testCase.NextTime,
            TotalSecs: testCase.TotalSecs,
            IsEmpty: testCase.IsEmpty,
            String: testCase.String,
        });
    }

    return testCases;
}

func TestScheduledTimeValidTestCases(test *testing.T) {
    for i, testCase := range getValidTestCases() {
        err := testCase.TimeSpec.Validate();
        if (err != nil) {
            test.Errorf("Case %d (%s) -- Failed to validate: '%s'.", i, testCase.ID, err);
            continue;
        }

        nextTime := testCase.TimeSpec.ComputeNextTime(baseTime);
        if (testCase.NextTime != nextTime) {
            test.Errorf("Case %d (%s) -- Incorrect next time. Expected: '%s', found: '%s'.", i, testCase.ID, testCase.NextTime, nextTime);
            continue;
        }

        totalSecs := testCase.TimeSpec.TotalSecs();
        if (testCase.TotalSecs != totalSecs) {
            test.Errorf("Case %d (%s) -- Incorrect total secs. Expected: %d, found: %d.", i, testCase.ID, testCase.TotalSecs, totalSecs);
            continue;
        }

        isEmpty := testCase.TimeSpec.IsEmpty();
        if (testCase.IsEmpty != isEmpty) {
            test.Errorf("Case %d (%s) -- Incorrect is empty. Expected: %v, found: %v.", i, testCase.ID, testCase.IsEmpty, isEmpty);
            continue;
        }

        stringVal := testCase.TimeSpec.String();
        if (testCase.String != stringVal) {
            test.Errorf("Case %d (%s) -- Incorrect string value. Expected: '%s', found: '%s'.", i, testCase.ID, testCase.String, stringVal);
            continue;
        }
    }
}

func TestScheduledTimeInvalidTestCases(test *testing.T) {
    for i, testCase := range invalidTestCases {
        err := testCase.Validate();
        if (err == nil) {
            test.Errorf("Case %d -- Validate failed to return an error on '%#v'.", i, testCase);
            continue;
        }
    }
}

type durationSpecTestCase struct {
    TimeSpec *durationSpec
    NextTime time.Time
    TotalSecs int
    IsEmpty bool
    String string
}

type timeOfDaySpecTestCase struct {
    TimeSpec timeOfDaySpec
    NextTime time.Time
    TotalSecs int
    IsEmpty bool
    String string
}

var validDurationCases []durationSpecTestCase = []durationSpecTestCase{
    durationSpecTestCase{
        TimeSpec: &durationSpec{0, 0, 0, 0},
        NextTime: time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC),
        TotalSecs: 0,
        IsEmpty: true,
        String: fmt.Sprintf("Every 0 days, 0 hours, 0 minutes, 0 seconds; (0 total seconds)"),
    },
    durationSpecTestCase{
        TimeSpec: &durationSpec{1, 0, 0, 0},
        NextTime: time.Date(2023, time.October, 2, 0, 0, 0, 0, time.UTC),
        TotalSecs: SECS_PER_DAY,
        IsEmpty: false,
        String: fmt.Sprintf("Every 1 days, 0 hours, 0 minutes, 0 seconds; (%d total seconds)", SECS_PER_DAY),
    },
    durationSpecTestCase{
        TimeSpec: &durationSpec{0, 1, 0, 0},
        NextTime: time.Date(2023, time.October, 1, 1, 0, 0, 0, time.UTC),
        TotalSecs: SECS_PER_HOUR,
        IsEmpty: false,
        String: fmt.Sprintf("Every 0 days, 1 hours, 0 minutes, 0 seconds; (%d total seconds)", SECS_PER_HOUR),
    },
    durationSpecTestCase{
        TimeSpec: &durationSpec{0, 0, 1, 0},
        NextTime: time.Date(2023, time.October, 1, 0, 1, 0, 0, time.UTC),
        TotalSecs: SECS_PER_MIN,
        IsEmpty: false,
        String: fmt.Sprintf("Every 0 days, 0 hours, 1 minutes, 0 seconds; (%d total seconds)", SECS_PER_MIN),
    },
    durationSpecTestCase{
        TimeSpec: &durationSpec{0, 0, 0, 1},
        NextTime: time.Date(2023, time.October, 1, 0, 0, 1, 0, time.UTC),
        TotalSecs: 1,
        IsEmpty: false,
        String: fmt.Sprintf("Every 0 days, 0 hours, 0 minutes, 1 seconds; (1 total seconds)"),
    },
    durationSpecTestCase{
        TimeSpec: &durationSpec{1, 2, 3, 4},
        NextTime: time.Date(2023, time.October, 2, 2, 3, 4, 0, time.UTC),
        TotalSecs: 1 * SECS_PER_DAY + 2 * SECS_PER_HOUR + 3 * SECS_PER_MIN + 4,
        IsEmpty: false,
        String: fmt.Sprintf("Every 1 days, 2 hours, 3 minutes, 4 seconds; (%d total seconds)", 1 * SECS_PER_DAY + 2 * SECS_PER_HOUR + 3 * SECS_PER_MIN + 4),
    },
};

var validTimeOfDayCases []timeOfDaySpecTestCase = []timeOfDaySpecTestCase{
    timeOfDaySpecTestCase{
        TimeSpec: timeOfDaySpec(""),
        NextTime: time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC),
        TotalSecs: SECS_PER_DAY,
        IsEmpty: true,
        String: fmt.Sprintf("Daily at 00:00:00."),
    },
    timeOfDaySpecTestCase{
        TimeSpec: timeOfDaySpec("00:00"),
        NextTime: time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC),
        TotalSecs: SECS_PER_DAY,
        IsEmpty: false,
        String: fmt.Sprintf("Daily at 00:00:00."),
    },
    timeOfDaySpecTestCase{
        TimeSpec: timeOfDaySpec("11:59"),
        NextTime: time.Date(2023, time.October, 1, 11, 59, 0, 0, time.UTC),
        TotalSecs: SECS_PER_DAY,
        IsEmpty: false,
        String: fmt.Sprintf("Daily at 11:59:00."),
    },
    timeOfDaySpecTestCase{
        TimeSpec: timeOfDaySpec("23:59"),
        NextTime: time.Date(2023, time.October, 1, 23, 59, 0, 0, time.UTC),
        TotalSecs: SECS_PER_DAY,
        IsEmpty: false,
        String: fmt.Sprintf("Daily at 23:59:00."),
    },
    timeOfDaySpecTestCase{
        TimeSpec: timeOfDaySpec("01:02"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 0, 0, time.UTC),
        TotalSecs: SECS_PER_DAY,
        IsEmpty: false,
        String: fmt.Sprintf("Daily at 01:02:00."),
    },
    timeOfDaySpecTestCase{
        TimeSpec: timeOfDaySpec("01:02:03"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 3, 0, time.UTC),
        TotalSecs: SECS_PER_DAY,
        IsEmpty: false,
        String: fmt.Sprintf("Daily at 01:02:03."),
    },
    timeOfDaySpecTestCase{
        TimeSpec: timeOfDaySpec("01:02:03.04"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 3, 4e7, time.UTC),
        TotalSecs: SECS_PER_DAY,
        IsEmpty: false,
        String: fmt.Sprintf("Daily at 01:02:03."),
    },
};

var invalidTestCases []timeSpec = []timeSpec{
    &durationSpec{-1, 0, 0, 0},
    &durationSpec{0, -2, 0, 0},
    &durationSpec{0, 0, -3, 0},
    &durationSpec{0, 0, 0, -4},
    &durationSpec{-4, -3, -2, -1},

    timeOfDaySpec("1"),
    timeOfDaySpec("1 PM"),
    timeOfDaySpec("1:00 PM"),
    timeOfDaySpec("01:02:03:04"),
    timeOfDaySpec("01:02.03"),
    timeOfDaySpec("24:01"),
    timeOfDaySpec("01:60"),
    timeOfDaySpec("01:02:60"),
};
