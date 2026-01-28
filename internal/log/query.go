package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edulinq/autograder/internal/timestamp"
)

const MARSHAL_ERROR string = "<error>"

// A representation of a query for server log records.
// A raw query can be processed (validated for structure and permissions)
// by the internal/procedures/log package.
// Because of some assumptions we make, log times before UNIX epoch are not supported.
// However, since this code was created well past that only time travelers should be concerned.
type RawLogQuery struct {
	LevelString  string `json:"level,omitempty"`
	AfterString  string `json:"after,omitempty"`
	PastString   string `json:"past,omitempty"`
	CourseID     string `json:"target-course,omitempty"`
	AssignmentID string `json:"target-assignment,omitempty"`
	TargetUser   string `json:"target-email,omitempty"`
}

// The fully parsed query to be executed.
// A full query is the conjunction of the present fields.
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

	parsed.CourseID, err = parseLogQueryID(this.CourseID)
	if err != nil {
		errs = append(errs, err)
	}

	parsed.AssignmentID, err = parseLogQueryID(this.AssignmentID)
	if err != nil {
		errs = append(errs, err)
	}

	// Assignment must have a course.
	if (parsed.AssignmentID != "") && (parsed.CourseID == "") {
		errs = append(errs, fmt.Errorf("Log queries with an assignment must also have a course."))
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

func parseLogQueryID(id string) (string, error) {
	if id == "" {
		return "", nil
	}

	return strings.TrimSpace(id), nil
}

func (this RawLogQuery) String() string {
	text, err := json.Marshal(this)
	if err != nil {
		return MARSHAL_ERROR
	}

	return string(text)
}

func (this ParsedLogQuery) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Level: '%s'", this.Level.String()))

	after := this.After.SafeString()
	if this.After.IsZero() {
		after = "< all time >"
	}
	builder.WriteString(fmt.Sprintf(", After: '%s'", after))

	course := this.CourseID
	if course == "" {
		course = "< all courses >"
	}
	builder.WriteString(fmt.Sprintf(", Course: '%s'", course))

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

func (this ParsedLogQuery) Match(record *Record) bool {
	if record == nil {
		return false
	}

	if record.Level < this.Level {
		return false
	}

	if (this.CourseID != "") && (record.Course != this.CourseID) {
		return false
	}

	// Assignment ID will only be matched on if the course ID also matches.
	courseMatch := ((this.CourseID != "") && (record.Course == this.CourseID))

	if (this.AssignmentID != "") && (!courseMatch || (record.Assignment != this.AssignmentID)) {
		return false
	}

	if (this.UserEmail != "") && (record.User != this.UserEmail) {
		return false
	}

	if record.Timestamp < this.After {
		return false
	}

	return true
}
