package log

import (
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
)

const ASSIGNMENT_ID = "exists"

func TestLogQueryBase(test *testing.T) {
	testCases := []struct {
		query       RawLogQuery
		expected    ParsedLogQuery
		errors      []string
		replaceTime bool
	}{
		{
			RawLogQuery{AfterString: "0"},
			ParsedLogQuery{},
			[]string{},
			false,
		},

		{
			RawLogQuery{
				LevelString: LEVEL_STRING_INFO,
			},
			ParsedLogQuery{},
			[]string{},
			false,
		},
		{
			RawLogQuery{
				LevelString: LEVEL_STRING_DEBUG,
			},
			ParsedLogQuery{
				Level: LevelDebug,
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
				After: timestamp.MustGuessFromString("2000-01-02T03:04:05Z"),
			},
			[]string{},
			false,
		},
		{
			RawLogQuery{
				AfterString: "2000-01-02",
			},
			ParsedLogQuery{
				After: timestamp.MustGuessFromString("2000-01-02T00:00:00Z"),
			},
			[]string{},
			false,
		},

		{
			RawLogQuery{
				PastString: "24h",
			},
			ParsedLogQuery{
				After: timestamp.MustGuessFromString("2000-01-02T03:04:05Z"),
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
				PastString:  "24h",
			},
			ParsedLogQuery{
				After: timestamp.MustGuessFromString("2000-01-02T03:04:05Z"),
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
		// Will be handled by validation (not parsing).
		{
			RawLogQuery{
				AssignmentID: "!!!",
			},
			ParsedLogQuery{
				AssignmentID: "!!!",
			},
			[]string{},
			false,
		},
		// Will be handled by validation (not parsing).
		{
			RawLogQuery{
				AssignmentID: "ZZZ",
			},
			ParsedLogQuery{
				AssignmentID: "ZZZ",
			},
			[]string{},
			false,
		},

		{
			RawLogQuery{
				TargetUser: "course-admin@test.edulinq.org",
			},
			ParsedLogQuery{
				UserEmail: "course-admin@test.edulinq.org",
			},
			[]string{},
			false,
		},
	}

	for i, testCase := range testCases {
		actual, errors := testCase.query.ParseStrings()

		if !reflect.DeepEqual(testCase.errors, errors) {
			test.Fatalf("Case %d: Errors not as expected. Expected: '%v', Actual: '%v'.", i,
				testCase.errors, errors)
		}

		if testCase.replaceTime {
			actual.After = testCase.expected.After
		}

		if testCase.expected != *actual {
			test.Fatalf("Case %d: Parsed query not as expected. Expected: '%v', Actual: '%v'.", i,
				testCase.expected, *actual)
		}
	}
}

func TestLogQueryParseTiming(test *testing.T) {
	now := timestamp.MustGuessFromString("2000-01-02T03:04:05Z")

	testCases := []struct {
		afterString string
		pastString  string
		expected    timestamp.Timestamp
	}{
		{"", "", timestamp.Zero()},
		{"2000-01-01T03:04:05Z", "", timestamp.MustGuessFromString("2000-01-01T03:04:05Z")},
		{"", "24h", timestamp.MustGuessFromString("2000-01-01T03:04:05Z")},
		{"2000-01-01T03:04:05Z", "1h", timestamp.MustGuessFromString("2000-01-02T02:04:05Z")},
	}

	for i, testCase := range testCases {
		actual, err := parseLogQueryTiming(now, testCase.afterString, testCase.pastString)
		if err != nil {
			test.Fatalf("Case %d: Unexpected error: '%v'.", i, err)
		}

		if testCase.expected != actual {
			test.Fatalf("Case %d: Time mismatch. Expected: '%s', Actual: '%s'.", i,
				testCase.expected.SafeString(), actual.SafeString())
		}
	}
}
