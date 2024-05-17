package model

import (
    "fmt"
    "testing"
    "time"

    "github.com/edulinq/autograder/internal/util"
)

// Test some dates to make sure that are marshed from JSON correctly.
func TestDates(test *testing.T) {
    testCases := []dateTestCase{
        dateTestCase{"2023-09-28T04:00:20+00:00", getTime("2023-09-28T04:00:20+00:00")},
        dateTestCase{"2023-09-28T04:00:20Z", getTime("2023-09-28T04:00:20Z")},
    };

    for i, testCase := range testCases {
        jsonString := fmt.Sprintf(`{"time": "%s"}`, testCase.Input);

        actual := make(map[string]time.Time);
        err := util.JSONFromString(jsonString, &actual);
        if (err != nil) {
            test.Fatal(err);
        }

        if (actual["time"] != testCase.Expected) {
            test.Errorf("Date case %d does not match. Expected '%s', Got '%s'.", i, testCase.Expected, actual["time"]);
        }
    }
}

type dateTestCase struct {
    Input string
    Expected time.Time
}

func getTime(value string) time.Time {
    result, err := time.Parse(time.RFC3339, value);
    if (err != nil) {
        panic(err);
    }

    return result;
}

func getSimpleTime(value string) time.Time {
    result, err := time.Parse(time.DateTime, value);
    if (err != nil) {
        panic(err);
    }

    return result;
}
