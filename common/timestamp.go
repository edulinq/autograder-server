package common

// Consistently handle time operations.
// Internally, all times should be a time.Time.
// Externally (including things internal, but serialized), all times should be considered a "timestamp" string.

import (
    "fmt"
    "time"

    "github.com/edulinq/autograder/log"
)

type Timestamp string;

const (
    TIMESTAMP_FORMAT = time.RFC3339;
    PRETTY_FORMAT = time.DateTime;
)

func NowTimestamp() Timestamp {
    return TimestampFromTime(time.Now());
}

func TimestampFromTime(instance time.Time) Timestamp {
    return Timestamp(instance.Format(TIMESTAMP_FORMAT));
}

func TimestampFromString(text string) (Timestamp, error) {
    instance, err := time.Parse(TIMESTAMP_FORMAT, text);
    if (err != nil) {
        return NowTimestamp(), fmt.Errorf("Failed to parse timestamp string '%s': '%w'.", text, err);
    }

    return TimestampFromTime(instance), nil;
}

func MustTimestampFromString(text string) Timestamp {
    instance, err := TimestampFromString(text);
    if (err != nil) {
        log.Fatal("Failed to parse timestamp text.", err, log.NewAttr("text", text), log.NewAttr("format", TIMESTAMP_FORMAT));
    }

    return instance;
}

func (this Timestamp) Validate() error {
    if (this.IsZero()) {
        return nil;
    }

    _, err := this.Time();
    if (err != nil) {
        return err;
    }

    return nil;
}

func (this Timestamp) IsZero() bool {
    return (string(this) == "");
}

func (this Timestamp) String() string {
    return string(this);
}

func (this Timestamp) ShouldPrettyString() string {
    instance, _ := this.Time();
    return instance.Format(PRETTY_FORMAT);
}

func (this Timestamp) Time() (time.Time, error) {
    instance, err := time.Parse(TIMESTAMP_FORMAT, string(this));
    if (err != nil) {
        return time.Time{}, fmt.Errorf("Failed to parse timestamp string '%s': '%w'.", string(this), err);
    }

    return instance, nil;
}

func (this Timestamp) MustTime() time.Time {
    instance, err := this.Time();
    if (err != nil) {
        log.Fatal("Failed to parse timestamp.", err, log.NewAttr("text", string(this)), log.NewAttr("format", TIMESTAMP_FORMAT));
    }

    return instance;
}
