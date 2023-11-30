package model

import (
    "testing"
    "time"
)

// 2023-10-01 00:00 Sunday
var baseTime time.Time = time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC)

type scheduledTimeTestCase struct {
    Time ScheduledTime
    NextTime time.Time
}

var testValidCases []scheduledTimeTestCase = []scheduledTimeTestCase{
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", ""),
        NextTime: time.Date(2023, time.October, 1, 0, 0, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "11:59"),
        NextTime: time.Date(2023, time.October, 1, 11, 59, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "23:59"),
        NextTime: time.Date(2023, time.October, 1, 23, 59, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "01:02"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "01:02:03"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 3, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "01:02:03.04"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 3, 4e7, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("Sunday", "01:02"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("SuNdaY", "01:02"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("Sun", "01:02"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("sun", "01:02"),
        NextTime: time.Date(2023, time.October, 1, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("mon", "01:02"),
        NextTime: time.Date(2023, time.October, 2, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("tue", "01:02"),
        NextTime: time.Date(2023, time.October, 3, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("wed", "01:02"),
        NextTime: time.Date(2023, time.October, 4, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("thu", "01:02"),
        NextTime: time.Date(2023, time.October, 5, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("fri", "01:02"),
        NextTime: time.Date(2023, time.October, 6, 1, 2, 0, 0, time.UTC),
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("sat", "01:02"),
        NextTime: time.Date(2023, time.October, 7, 1, 2, 0, 0, time.UTC),
    },
};

func TestScheduledTimeValidTestCases(test *testing.T) {
    for i, testCase := range testValidCases {
        err := testCase.Time.Validate();
        if (err != nil) {
            test.Errorf("Test Case %d -- Failed to validate: '%s'.", i, err);
            continue;
        }

        nextTime := testCase.Time.computeNextTime(baseTime);
        if (nextTime != testCase.NextTime) {
            test.Errorf("Test Case %d -- Expected next time '%s', found next time '%s.", i, testCase.NextTime, nextTime);
            continue;
        }
    }
}

var testInvalidCases []scheduledTimeTestCase = []scheduledTimeTestCase{
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("ZZZ", ""),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("Scrubday", ""),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("m0n", ""),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "1"),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "1 PM"),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "1:00 PM"),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "01:02:03:04"),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "01:02.03"),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "24:01"),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "01:60"),
        NextTime: time.Time{},
    },
    scheduledTimeTestCase{
        Time: newScheduledTimeForTestiong("", "01:02:60"),
        NextTime: time.Time{},
    },
};

func TestScheduledTimeInvalidTestCases(test *testing.T) {
    for i, testCase := range testInvalidCases {
        err := testCase.Time.Validate();
        if (err == nil) {
            test.Errorf("Test Case %d -- Validate failed to return an error on '%s'.", i, testCase.Time.String());
            continue;
        }
    }
}
