package common

import (
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/edulinq/autograder/log"
)

type RawLogQuery struct {
    LevelString string `json:"level"`
    AfterString string `json:"after"`
    AssignmentID string `json:"assignment-id"`
    TargetUser string `json:"target-email"`
}

type ParsedLogQuery struct {
    Level log.LogLevel
    After time.Time
    AssignmentID string
    UserID string
}

type courseInferface interface {
    HasAssignment(id string) bool
}

// Like Parse(), but return strings instead of errors.
// The string slice will always be non-nil, but only non-empty when an error occurred.
func (this RawLogQuery) ParseStrings(course courseInferface) (*ParsedLogQuery, []string) {
    messages := make([]string, 0, 0);

    parsed, errs := this.Parse(course);
    if (errs != nil) {
        for _, err := range errs {
            messages = append(messages, err.Error());
        }
    }

    return parsed, messages;
}

// Like Parse(), but return a single joined error.
func (this RawLogQuery) ParseJoin(course courseInferface) (*ParsedLogQuery, error) {
    var joinedError error = nil;

    parsed, errs := this.Parse(course);
    if (errs != nil) {
        for _, err := range errs {
            joinedError = errors.Join(joinedError, err);
        }
    }

    return parsed, joinedError;
}

// Parse a raw log query and return a version with clean attributes.
// These attributes have not been validated against any database, but have been parsed from strings.
// The returned collection of errors will only be non-nil if errors occurred.
func (this RawLogQuery) Parse(course courseInferface) (*ParsedLogQuery, []error) {
    var parsed ParsedLogQuery;

    var err error;
    var errs []error = nil;

    parsed.Level, err = ParseLogQueryLevel(this.LevelString);
    if (err != nil) {
        errs = append(errs, err);
    }

    parsed.After, err = ParseLogQueryAfter(this.AfterString);
    if (err != nil) {
        errs = append(errs, err);
    }

    parsed.AssignmentID, err = ParseLogQueryAssignmentID(this.AssignmentID, course);
    if (err != nil) {
        errs = append(errs, err);
    }

    parsed.UserID = this.TargetUser;

    return &parsed, errs;
}

func (this ParsedLogQuery) String() string {
    var builder strings.Builder;

    builder.WriteString(fmt.Sprintf("Level: '%s'", this.Level.String()));

    after := TimestampFromTime(this.After).String();
    if (this.After.IsZero()) {
        after = "< all time >";
    }
    builder.WriteString(fmt.Sprintf(", After: '%s'", after));

    assignment := this.AssignmentID;
    if (assignment == "") {
        assignment = "< all assignments >";
    }
    builder.WriteString(fmt.Sprintf(", Assignment: '%s'", assignment));

    user := this.UserID;
    if (user == "") {
        user = "< all users >";
    }
    builder.WriteString(fmt.Sprintf(", User: '%s'", user));

    return builder.String();
}

func ParseLogQueryLevel(levelString string) (log.LogLevel, error) {
    if (levelString == "") {
        return log.LevelInfo, nil;
    }

    level, err := log.ParseLevel(levelString);
    if (err != nil) {
        return log.LevelInfo, fmt.Errorf("Could not parse 'level' componet of log query ('%s'): '%v'.", levelString, err);
    }

    return level, nil;
}

func ParseLogQueryAfter(afterString string) (time.Time, error) {
    if (afterString == "") {
        return time.Time{}, nil;
    }

    timestamp, err := TimestampFromString(afterString);
    if (err != nil) {
        return time.Time{}, fmt.Errorf("Could not parse 'after' componet of log query ('%s'): '%v'.", afterString, err);
    }

    after, err := timestamp.Time();
    if (err != nil) {
        return time.Time{}, fmt.Errorf("Could not extract time from 'after' componet of log query ('%s'): '%v'.", afterString, err);
    }

    return after, nil;
}

func ParseLogQueryAssignmentID(assignmentString string, course courseInferface) (string, error) {
    if (assignmentString == "") {
        return "", nil;
    }

    assignmentID, err := ValidateID(assignmentString);
    if (err != nil) {
        return "", fmt.Errorf("Could not parse 'assignment' componet of log query ('%s'): '%v'.", assignmentString, err);
    }

    if (!course.HasAssignment(assignmentID)) {
        return "", fmt.Errorf("Unknown assignment given for 'assignment' componet of log query ('%s').", assignmentString);
    }

    return assignmentID, nil;
}
