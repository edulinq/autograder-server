package timestamp

// Consistently handle time operations.
// All times representing an instance in time (datetime) should be represented with a timestamp.Timestamp.
// A timestamp is the number of milliseconds (int64) since the UNIX epoch (which is in UTC).
// Time resolutions smaller than a millisecond should be handled by packages internally.
// Users can convert to local time for display purposes,
// but any time passed between packages or serialized should be a timestamp.Timestamp.

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// A safe (always valid) time representation.
// A timestamp is the number of milliseconds (int64) since the UNIX epoch (which is in UTC).
type Timestamp int64

const (
	STRING_FORMAT_SAFE   = time.RFC3339Nano
	STRING_FORMAT_UNSAFE = time.DateTime

	// Guessing thresholds.
	UNIXTIME_THRESHOLD_SECS  = 1e10
	UNIXTIME_THRESHOLD_MSECS = 1e13
	UNIXTIME_THRESHOLD_USECS = 1e16

	// Conversions
	MSECS_PER_SECS  = 1000
	MSECS_PER_MINS  = MSECS_PER_SECS * 60
	MSECS_PER_HOURS = MSECS_PER_MINS * 60
	MSECS_PER_DAYS  = MSECS_PER_HOURS * 24
)

// The formats to try when guessing a timestamp (in order).
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
	time.DateTime,
	time.DateOnly,
}

func Now() Timestamp {
	return FromGoTime(time.Now())
}

func Zero() Timestamp {
	return Timestamp(0)
}

func ZeroPointer() *Timestamp {
	value := Zero()
	return &value
}

func FromGoTime(instance time.Time) Timestamp {
	return Timestamp(instance.UnixMilli())
}

func FromGoTimePointer(instance *time.Time) *Timestamp {
	if instance == nil {
		return nil
	}

	timestamp := Timestamp(instance.UnixMilli())
	return &timestamp
}

func FromGoTimeDuration(duration time.Duration) Timestamp {
	return Timestamp(duration / time.Millisecond)
}

func FromMSecs(msecs int64) Timestamp {
	return Timestamp(msecs)
}

// Try to parse a time out of a string.
// This is not efficient or robust, and should not be used in any non-interactive environments.
// Purely digit strings will be converted to ints and treated as Unix times.
// Other strings will be attempted to be parsed with several different time formats,
// the first format that does not throw an error will be used.
func GuessFromString(text string) (Timestamp, error) {
	text = strings.TrimSpace(text)

	// Check for an int (unix time).
	unixTime, err := strconv.ParseInt(text, 10, 64)
	if err == nil {
		// Use reasonable thresholds to guess the units of the value (sec, msec, usec, nsec).
		if unixTime < UNIXTIME_THRESHOLD_SECS {
			// Time is in secs.
			return Timestamp(unixTime * 1000), nil
		} else if unixTime < UNIXTIME_THRESHOLD_MSECS {
			// Time is in msecs.
			return Timestamp(unixTime), nil
		} else if unixTime < UNIXTIME_THRESHOLD_USECS {
			// Time is in usecs.
			return Timestamp(unixTime / 1000), nil
		} else {
			// Time is in nsecs.
			return Timestamp(unixTime / 1000 / 1000), nil
		}
	}

	// Try all the formats, and stop at the first non-error one.
	for _, format := range timeFormats {
		instance, err := time.ParseInLocation(format, text, time.Local)
		if err == nil {
			return FromGoTime(instance), nil
		}
	}

	return Zero(), fmt.Errorf("Could not guess time '%s'.", text)
}

func MustGuessFromString(text string) Timestamp {
	timestamp, err := GuessFromString(text)
	if err != nil {
		panic(fmt.Sprintf("Failed to guess timestamp from string ('%s'): '%v'.", text, err))
	}

	return timestamp
}

func (this Timestamp) IsZero() bool {
	return (this == 0)
}

func (this Timestamp) ToGoTime() time.Time {
	return time.UnixMilli(int64(this))
}

func (this Timestamp) ToMSecs() int64 {
	return int64(this)
}

func (this Timestamp) ToSecs() float64 {
	return float64(this) / MSECS_PER_SECS
}

func (this Timestamp) ToMins() float64 {
	return float64(this) / MSECS_PER_MINS
}

func (this Timestamp) ToHours() float64 {
	return float64(this) / MSECS_PER_HOURS
}

func (this Timestamp) ToDays() float64 {
	return float64(this) / MSECS_PER_DAYS
}

func (this Timestamp) ToGoTimeDuration() time.Duration {
	return time.Duration(int64(this) * int64(time.Millisecond))
}

func (this *Timestamp) SafeString() string {
	if this == nil {
		return "<nil>"
	}

	return this.ToGoTime().Format(STRING_FORMAT_SAFE)
}

func (this Timestamp) UnsafePrettyString() string {
	return this.ToGoTime().Format(STRING_FORMAT_UNSAFE)
}

// Convert the timestamp to a string format meant to be embedded (and later retrieved via regex) into a string message.
// This format is readable by a human (if they understand epochs) and by a machine (via regex).
// Format: /<timestamp:(-?\d+|nil)>/.
func (this *Timestamp) SafeMessage() string {
	if this == nil {
		return "<timestamp:nil>"
	}

	return fmt.Sprintf("<timestamp:%d>", int64(*this))
}
