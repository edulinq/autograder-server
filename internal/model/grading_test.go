package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/util"
)

// Test some dates to make sure that are marshalled from JSON correctly.
func TestDates(test *testing.T) {
	testCases := []dateTestCase{
		dateTestCase{"2023-09-28T04:00:20+00:00", getTime("2023-09-28T04:00:20+00:00")},
		dateTestCase{"2023-09-28T04:00:20Z", getTime("2023-09-28T04:00:20Z")},
	}

	for i, testCase := range testCases {
		jsonString := fmt.Sprintf(`{"time": "%s"}`, testCase.Input)

		actual := make(map[string]time.Time)
		err := util.JSONFromString(jsonString, &actual)
		if err != nil {
			test.Fatal(err)
		}

		if actual["time"] != testCase.Expected {
			test.Errorf("Date case %d does not match. Expected '%s', Got '%s'.", i, testCase.Expected, actual["time"])
		}
	}
}

type dateTestCase struct {
	Input    string
	Expected time.Time
}

func getTime(value string) time.Time {
	result, err := time.ParseInLocation(time.RFC3339, value, time.Local)
	if err != nil {
		panic(err)
	}

	return result
}

func getSimpleTime(value string) time.Time {
	result, err := time.ParseInLocation(time.DateTime, value, time.Local)
	if err != nil {
		panic(err)
	}

	return result
}

// Test hard fail and skipped are marshalled from JSON correctly.
func TestHardFailAndSkipped(test *testing.T) {
	testCases := []struct {
		Score    float64
		HardFail bool
		Skipped  bool
	}{
		{1, false, false},
		{0, false, true},
		{0, true, false},
		{0, true, true},
	}

	for i, testCase := range testCases {
		expectedInfo := GradingInfo{
			ID:           "test-info-123",
			ShortID:      "123",
			CourseID:     "test-course",
			AssignmentID: "test-assignment",
			User:         "course-student",
			Message:      "Great job!",

			Questions: []*GradedQuestion{
				&GradedQuestion{
					Name:      "Q1",
					MaxPoints: 1,
					Score:     testCase.Score,
					HardFail:  testCase.HardFail,
					Skipped:   testCase.Skipped,
				},
			},
		}

		expectedInfo.ComputePoints()

		actualInfoJson := util.MustToJSON(expectedInfo)

		var actualInfo GradingInfo
		util.MustJSONFromString(actualInfoJson, &actualInfo)

		if !expectedInfo.Equals(actualInfo, true) {
			test.Errorf("Case %d: Unexpected result. Expected info '%v', actual info '%v'.",
				i, util.MustToJSONIndent(expectedInfo), util.MustToJSONIndent(actualInfo))
		}
	}
}
