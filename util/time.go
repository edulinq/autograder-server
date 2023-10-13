package util

// Consistently handle time operations.
// Externally, all times should be considered a "timestamp" string.
// Internally, all times should be a time.Time.

import (
    "time"
)

const TIMESTAMP_FORMAT = time.RFC3339;

func NowTimestamp() string {
    return ToTimestamp(time.Now());
}

func ToTimestamp(instance time.Time) string {
    return instance.Format(TIMESTAMP_FORMAT);
}

func FromTimestamp(timestamp string) (time.Time, error) {
    return time.Parse(TIMESTAMP_FORMAT, timestamp);
}
