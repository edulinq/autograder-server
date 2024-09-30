package timestamp

import (
	"testing"
	"time"
)

func TestTimestamp(test *testing.T) {
	testCases := []struct {
		value       int64
		stringValue string
		expected    Timestamp
	}{
		{0, "1970-01-01T00:00:00Z", Zero()},
		{1, "1970-01-01T00:00:00.001000000Z", Timestamp(1)},
		{-1, "1969-12-31T23:59:59.999000000Z", Timestamp(-1)},

		// Resolution too small.
		{0, "1970-01-01T00:00:00.000999999Z", Zero()},

		// Go's time.Duration is in nsecs.
		{0, "1970-01-01T00:00:00.000000000Z", FromGoTimeDuration(time.Nanosecond * 0)},
		{0, "1970-01-01T00:00:00.000000000Z", FromGoTimeDuration(time.Nanosecond * 1)},
		{1, "1970-01-01T00:00:00.001000000Z", FromGoTimeDuration(time.Nanosecond * 1 * time.Millisecond)},
	}

	for i, testCase := range testCases {
		// Convert from int directly.
		actualFromInt := Timestamp(testCase.value)

		// Parse string.
		actualFromString, err := GuessFromString(testCase.stringValue)
		if err != nil {
			test.Errorf("Case %d: Failed to get timestamp from string '%s': '%v'.", i, testCase.stringValue, err)
			continue
		}

		// Ensure int and string representations match.
		if actualFromInt != actualFromString {
			test.Errorf("Case %d: Int (%v) and string (%v) values do not match.", i, actualFromInt, actualFromString)
			continue
		}

		actual := actualFromInt

		// Ensure we have the value we expect.
		if actual != testCase.expected {
			test.Errorf("Case %d: Timestamp not as expected. Actual: %v, Expected: %v.", i, actual, testCase.expected)
			continue
		}

		// Ensure that conversions to-from Go time work.
		actualFromGoTime := FromGoTime(actual.ToGoTime())
		if actual != actualFromGoTime {
			test.Errorf("Case %d: Actual (%v) and Go time (%v) values do not match.", i, actual, actualFromGoTime)
			continue
		}

		// Ensure that safe string conversions work.
		actualFromSafeString, err := GuessFromString(actual.SafeString())
		if err != nil {
			test.Errorf("Case %d: Failed to get timestamp from safe string '%s': '%v'.", i, actual.SafeString(), err)
			continue
		}

		if actual != actualFromSafeString {
			test.Errorf("Case %d: Actual (%v) and safe string (%v) values do not match.", i, actual, actualFromSafeString)
			continue
		}
	}
}

func TestTimestampGuessTime(test *testing.T) {
	testCases := []struct {
		value    string
		expected Timestamp
		isError  bool
	}{
		// All these were generated from Timestamp(1),
		// but most lost resolution in the format.
		{"1970-01-01T00:00:00.001Z", Timestamp(1), false},         // time.RFC3339Nano
		{"1970-01-01T00:00:00Z", Timestamp(0), false},             // time.RFC3339
		{"Thu, 01 Jan 1970 00:00:00 +0000", Timestamp(0), false},  // time.RFC1123Z
		{"Thu, 01 Jan 1970 00:00:00 UTC", Timestamp(0), false},    // time.RFC1123
		{"Thursday, 01-Jan-70 00:00:00 UTC", Timestamp(0), false}, // time.RFC850
		{"01 Jan 70 00:00 +0000", Timestamp(0), false},            // time.RFC822Z
		{"Thu Jan 01 00:00:00 +0000 1970", Timestamp(0), false},   // time.RubyDate
		{"Thu Jan  1 00:00:00 UTC 1970", Timestamp(0), false},     // time.UnixDate
		{"Thu Jan  1 00:00:00 1970", Timestamp(0), false},         // time.ANSIC
		{"01 Jan 70 00:00 UTC", Timestamp(0), false},              // time.RFC822
		{"01/01 12:00:00AM '70 +0000", Timestamp(0), false},       // time.Layout
		{"1970-01-01 00:00:00", Timestamp(0), false},              // time.DateTime
		{"1970-01-01", Timestamp(0), false},                       // time.DateOnly
	}

	for i, testCase := range testCases {
		actual, err := GuessFromString(testCase.value)
		if err != nil {
			if testCase.isError {
				continue
			}

			test.Errorf("Case %d: Failed to get timestamp from string '%s': '%v'.", i, testCase.value, err)
			continue
		}

		if testCase.isError {
			test.Errorf("Case %d: Failed to generate an error when one was expected: '%s'.", i, testCase.value)
			continue
		}

		if actual != testCase.expected {
			test.Errorf("Case %d: Timestamp not as expected. Actual: %v, Expected: %v.", i, actual, testCase.expected)
			continue
		}
	}
}

func TestTimestampMessage(test *testing.T) {
	testCases := []struct {
		input    *Timestamp
		expected string
	}{
		{nil, "<timestamp:nil>"},
		{newTimestampPointer(0), "<timestamp:0>"},
		{newTimestampPointer(1), "<timestamp:1>"},
		{newTimestampPointer(-1), "<timestamp:-1>"},
	}

	for i, testCase := range testCases {
		actual := testCase.input.SafeMessage()
		if testCase.expected != actual {
			test.Errorf("Case %d: Message not as expected. Actual: %v, Expected: %v.", i, actual, testCase.expected)
			continue
		}
	}
}

func newTimestampPointer(value int64) *Timestamp {
	timestamp := Timestamp(value)
	return &timestamp
}

func TestTimestampComparisons(test *testing.T) {
	smaller := Timestamp(0)
	sameAsSmaller := Zero()
	larger := Timestamp(1)

	if smaller > larger {
		test.Fatalf("> failed.")
	}

	if smaller >= larger {
		test.Fatalf(">= failed.")
	}

	if smaller != sameAsSmaller {
		test.Fatalf("!= failed.")
	}

	if !(smaller == sameAsSmaller) {
		test.Fatalf("== failed.")
	}

	if larger <= smaller {
		test.Fatalf("<= failed.")
	}

	if larger < smaller {
		test.Fatalf("< failed.")
	}
}
