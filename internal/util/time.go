package util

import (
    "fmt"
    "strconv"
    "strings"
    "time"
)

const (
    UNIXTIME_THRESHOLD_SECS = 1e10
    UNIXTIME_THRESHOLD_MSECS = 1e13
    UNIXTIME_THRESHOLD_USECS = 1e16
)

// The formats to try when guessing a time (in order).
var timeFormats []string = []string{
    time.RFC3339Nano,
    time.RFC3339,
    time.RFC1123Z,
    time.RFC1123,
    time.RFC850,
    time.RFC822Z,
    time.RubyDate,
    time.UnixDate,
    time.ANSIC,
    time.RFC822,
    time.Layout,
    time.StampNano,
    time.StampMicro,
    time.StampMilli,
    time.Stamp,
    time.DateTime,
    time.DateOnly,
    time.TimeOnly,
};

// Try to parse a time out of a string.
// This is not efficient or robust, and should not be used in any non-interactive environments.
// Purely digit strings will be converted to ints and treated as Unix times.
// Other strings will be attempted to be parsed with several different time formats,
// the first format that does not throw an error will be used.
func GuessTime(text string) (time.Time, error) {
    text = strings.TrimSpace(text);

    // Check for an int (unix time).
    unixTime, err := strconv.ParseInt(text, 10, 64);
    if (err == nil) {
        // Use reasonable thresholds to guess the units of the value (sec, msec, usec, nsec).
        if (unixTime < UNIXTIME_THRESHOLD_SECS) {
            return time.Unix(unixTime, 0), nil;
        } else if (unixTime < UNIXTIME_THRESHOLD_MSECS) {
            return time.UnixMilli(unixTime), nil;
        } else if (unixTime < UNIXTIME_THRESHOLD_USECS) {
            return time.UnixMicro(unixTime), nil;
        } else {
            return time.Unix(0, unixTime), nil;
        }
    }

    // Try all the formats, and stop at the first non-error one.
    for _, format := range timeFormats {
        instance, err := time.Parse(format, text);
        if (err == nil) {
            return instance, nil;
        }
    }

    return time.Time{}, fmt.Errorf("Could not guess time '%s'.", text);
}
