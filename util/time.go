package util

// Consistently handle time operations.
// Externally, all times should be considered a "timestamp" string.
// Internally, all times should be a time.Time.

// TEST - Remove, replace with common/time

import (
    "time"

    "github.com/rs/zerolog/log"
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

func MustFromTimestamp(timestamp string) time.Time {
    value, err := FromTimestamp(timestamp);
    if (err != nil) {
        log.Fatal().Err(err).Str("timestamp", timestamp).Str("format", TIMESTAMP_FORMAT).
                Msg("Failed to parse timestamp.");
    }

    return value;
}
