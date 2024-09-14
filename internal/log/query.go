package log

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edulinq/autograder/internal/timestamp"
)

// A representation of a query for server log records.
// A raw query can be processed (validated for structure and permissions)
// by the internal/procedures/log package.
// Because of some assumptions we make, log times before UNIX epoch are not supported.
// However, since this code was created well past that only time travelers should be concerned.
type RawLogQuery struct {
	LevelString  string `json:"level"`
	AfterString  string `json:"after"`
	PastString   string `json:"past"`
	AssignmentID string `json:"target-assignment"`
	CourseID     string `json:"target-course"`
	TargetUser   string `json:"target-email"`
}

type ParsedLogQuery struct {
	Level        LogLevel
	After        timestamp.Timestamp
	CourseID     string
	AssignmentID string
	UserEmail    string
}

// Parse a raw log query and return a version with clean attributes.
// These attributes have not been validated against any database, but have been parsed from strings.
// The returned collection of errors will only be non-nil if errors occurred.
func (this RawLogQuery) Parse() (*ParsedLogQuery, []error) {
	var parsed ParsedLogQuery

	var err error = nil
	var errs []error = nil

	parsed.Level, err = parseLogQueryLevel(this.LevelString)
	if err != nil {
		errs = append(errs, err)
	}

	parsed.After, err = parseLogQueryTiming(timestamp.Now(), this.AfterString, this.PastString)
	if err != nil {
		errs = append(errs, err)
	}

	parsed.AssignmentID, err = parseLogQueryAssignmentID(this.AssignmentID)
	if err != nil {
		errs = append(errs, err)
	}

	parsed.UserEmail = this.TargetUser

	return &parsed, errs
}

// Like Parse(), but return strings instead of errors.
// The string slice will always be non-nil, but only non-empty when an error occurred.
func (this RawLogQuery) ParseStrings() (*ParsedLogQuery, []string) {
	messages := make([]string, 0, 0)

	parsed, errs := this.Parse()
	if errs != nil {
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
	}

	return parsed, messages
}

// Like Parse(), but return a single joined error.
func (this RawLogQuery) ParseJoin() (*ParsedLogQuery, error) {
	var joinedError error = nil

	parsed, errs := this.Parse()
	if errs != nil {
		for _, err := range errs {
			joinedError = errors.Join(joinedError, err)
		}
	}

	return parsed, joinedError
}

func parseLogQueryLevel(levelString string) (LogLevel, error) {
	if levelString == "" {
		return LevelInfo, nil
	}

	level, err := ParseLevel(levelString)
	if err != nil {
		return LevelInfo, fmt.Errorf("Could not parse 'level' component of log query ('%s'): '%v'.", levelString, err)
	}

	return level, nil
}

// Parse both the after and past, and then resolve that the actual after time should be.
// If both after and past are given, the latter of the resulting times will be returned.
// If neither is given, the same time as |now| will be returned.
func parseLogQueryTiming(now timestamp.Timestamp, afterString string, pastString string) (timestamp.Timestamp, error) {
	afterTime, err := parseLogQueryAfter(afterString)
	if err != nil {
		return timestamp.Zero(), err
	}

	if pastString == "" {
		return afterTime, nil
	}

	pastDuration, err := parseLogQueryPastDuration(pastString)
	if err != nil {
		return timestamp.Zero(), err
	}

	pastTime := now - pastDuration

	if afterString == "" {
		return pastTime, nil
	}

	// Return the latter of the two times.
	// Note that both times could have not been provided,
	// in which case they will have zero values.
	// Then (now - 0) will be the chosen time.
	if afterTime > pastTime {
		return afterTime, nil
	}

	return pastTime, nil
}

func parseLogQueryAfter(afterString string) (timestamp.Timestamp, error) {
	if afterString == "" {
		return timestamp.Zero(), nil
	}

	after, err := timestamp.GuessFromString(afterString)
	if err != nil {
		return timestamp.Zero(), fmt.Errorf("Could not parse 'after' component of log query ('%s'): '%v'.", afterString, err)
	}

	return after, nil
}

func parseLogQueryPastDuration(pastString string) (timestamp.Timestamp, error) {
	if pastString == "" {
		return timestamp.Zero(), nil
	}

	pastDuration, err := time.ParseDuration(pastString)
	if err != nil {
		return timestamp.Zero(), fmt.Errorf("Could not parse 'past' component of log query ('%s'): '%v'.", pastString, err)
	}

	if pastDuration < 0 {
		return timestamp.Zero(), fmt.Errorf("Negative duration given for 'past' component of log query ('%s').", pastString)
	}

	return timestamp.FromGoTimeDuration(pastDuration), nil
}

func parseLogQueryAssignmentID(assignmentString string) (string, error) {
	if assignmentString == "" {
		return "", nil
	}

	return strings.TrimSpace(assignmentString), nil
}

func (this ParsedLogQuery) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Level: '%s'", this.Level.String()))

	after := this.After.SafeString()
	if this.After.IsZero() {
		after = "< all time >"
	}
	builder.WriteString(fmt.Sprintf(", After: '%s'", after))

	assignment := this.AssignmentID
	if assignment == "" {
		assignment = "< all assignments >"
	}
	builder.WriteString(fmt.Sprintf(", Assignment: '%s'", assignment))

	user := this.UserEmail
	if user == "" {
		user = "< all users >"
	}
	builder.WriteString(fmt.Sprintf(", User: '%s'", user))

	return builder.String()
}
