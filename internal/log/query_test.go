package log

import (
	"reflect"
	"testing"
	"time"

	"github.com/edulinq/autograder/internal/timestamp"
)

const (
	COURSE_ID     = "exists-course"
	ASSIGNMENT_ID = "exists-assignment"
)

func TestLogQueryBase(test *testing.T) {
	// Force the local time to UTC for tests.
	oldLocal := time.Local
	time.Local = time.UTC
	defer func() {
		time.Local = oldLocal
	}()

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
				CourseID: COURSE_ID,
			},
			ParsedLogQuery{
				CourseID: COURSE_ID,
			},
			[]string{},
			false,
		},
		// Will be handled by validation (not parsing).
		{
			RawLogQuery{
				CourseID: "!!!",
			},
			ParsedLogQuery{
				CourseID: "!!!",
			},
			[]string{},
			false,
		},
		// Will be handled by validation (not parsing).
		{
			RawLogQuery{
				CourseID: "ZZZ",
			},
			ParsedLogQuery{
				CourseID: "ZZZ",
			},
			[]string{},
			false,
		},

		{
			RawLogQuery{
				CourseID:     COURSE_ID,
				AssignmentID: ASSIGNMENT_ID,
			},
			ParsedLogQuery{
				CourseID:     COURSE_ID,
				AssignmentID: ASSIGNMENT_ID,
			},
			[]string{},
			false,
		},
		// Will be handled by validation (not parsing).
		{
			RawLogQuery{
				CourseID:     COURSE_ID,
				AssignmentID: "!!!",
			},
			ParsedLogQuery{
				CourseID:     COURSE_ID,
				AssignmentID: "!!!",
			},
			[]string{},
			false,
		},
		// Will be handled by validation (not parsing).
		{
			RawLogQuery{
				CourseID:     COURSE_ID,
				AssignmentID: "ZZZ",
			},
			ParsedLogQuery{
				CourseID:     COURSE_ID,
				AssignmentID: "ZZZ",
			},
			[]string{},
			false,
		},

		// Assignments must have a course.
		{
			RawLogQuery{
				AssignmentID: ASSIGNMENT_ID,
			},
			ParsedLogQuery{},
			[]string{
				"Log queries with an assignment must also have a course.",
			},
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
		// Ensure raw queries can cleanly convert to JSON.
		if testCase.query.String() == MARSHAL_ERROR {
			test.Errorf("Case %d: Could not marshal query.", i)
			continue
		}

		actual, errors := testCase.query.ParseStrings()

		if !reflect.DeepEqual(testCase.errors, errors) {
			test.Errorf("Case %d: Errors not as expected. Expected: '%v', Actual: '%v'.", i, testCase.errors, errors)
			continue
		}

		if len(testCase.errors) > 0 {
			continue
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
			test.Errorf("Case %d: Unexpected error: '%v'.", i, err)
			continue
		}

		if testCase.expected != actual {
			test.Errorf("Case %d: Time mismatch. Expected: '%s', Actual: '%s'.", i,
				testCase.expected.SafeString(), actual.SafeString())
			continue
		}
	}
}

func TestParsedLogQueryMatch(test *testing.T) {
	testCases := []struct {
		query    ParsedLogQuery
		record   *Record
		expected bool
	}{
		// Empty
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "",
				AssignmentID: "",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Now(),
				Course:     "",
				Assignment: "",
				User:       "",
			},
			true,
		},
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "",
				AssignmentID: "",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			true,
		},

		// Level
		{
			ParsedLogQuery{
				Level:        LevelDebug,
				After:        timestamp.Timestamp(0),
				CourseID:     "",
				AssignmentID: "",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			true,
		},
		{
			ParsedLogQuery{
				Level:        LevelWarn,
				After:        timestamp.Timestamp(0),
				CourseID:     "",
				AssignmentID: "",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			false,
		},

		// Course
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "course101",
				AssignmentID: "",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			true,
		},
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "ZZZ",
				AssignmentID: "",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			false,
		},

		// Assignment
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "course101",
				AssignmentID: "hw0",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			true,
		},
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "course101",
				AssignmentID: "ZZZ",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			false,
		},
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "",
				AssignmentID: "hw0",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			false,
		},

		// User
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "",
				AssignmentID: "",
				UserEmail:    "course-student@test.edulinq.org",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			true,
		},
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "",
				AssignmentID: "",
				UserEmail:    "ZZZ@test.edulinq.org",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			false,
		},

		// After
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(0),
				CourseID:     "",
				AssignmentID: "",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			true,
		},
		{
			ParsedLogQuery{
				Level:        LevelInfo,
				After:        timestamp.Timestamp(200),
				CourseID:     "",
				AssignmentID: "",
				UserEmail:    "",
			},
			&Record{
				Level:      LevelInfo,
				Timestamp:  timestamp.Timestamp(100),
				Course:     "course101",
				Assignment: "hw0",
				User:       "course-student@test.edulinq.org",
			},
			false,
		},
	}

	for i, testCase := range testCases {
		actual := testCase.query.Match(testCase.record)

		if testCase.expected != actual {
			test.Errorf("Case %d: Mismatch. Expected: '%v', Actual: '%v'.", i, testCase.expected, actual)
			continue
		}
	}
}
