package common

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/edulinq/autograder/internal/log"
)

const (
	TIME_LAYOUT_MINS = "15:04"
	TIME_LAYOUT_SECS = "15:04:05"

	// The duration's nanoseconds need to fit iside an int64.
	// To ensure it fits without overflowing, limit the size of each component.
	// Technically, we could not limit them and just check for a negative number of nanoseconds,
	// but that does not protect against multiple overflows.
	MAX_NSECS = int64(math.MaxInt64) / 10

	MAX_USECS = MAX_NSECS / int64(time.Microsecond)
	MAX_SECS  = MAX_NSECS / int64(time.Second)
	MAX_MINS  = MAX_NSECS / int64(time.Minute)
	MAX_HOURS = MAX_NSECS / int64(time.Hour)
	MAX_DAYS  = MAX_NSECS / (24 * int64(time.Hour))
)

// This struct should always have Validate() called after construction.
// All other methods will assume Validate() returns no error.
type ScheduledTime struct {
	Every DurationSpec  `json:"every,omitempty"`
	Daily TimeOfDaySpec `json:"daily,omitempty"`
}

type timeSpec interface {
	Validate() error
	TotalNanosecs() int64
	IsEmpty() bool
	// Get the next time to run (starting at the passed in time).
	ComputeNextTime(startTime time.Time) time.Time
	String() string
}

type TimeOfDaySpec string

type DurationSpec struct {
	Days    int64 `json:"days,omitempty"`
	Hours   int64 `json:"hours,omitempty"`
	Minutes int64 `json:"minutes,omitempty"`
	Seconds int64 `json:"seconds,omitempty"`

	// Microseconds in not exposed in JSON and in meant for testing.
	Microseconds int64 `json:"-"`
}

func (this *DurationSpec) Validate() error {
	if this.Days < 0 {
		return fmt.Errorf("Duration cannot have negative days, found '%d'.", this.Days)
	}

	if this.Days >= MAX_DAYS {
		return fmt.Errorf("Duration has too many days (%d), max: %d.", this.Days, MAX_DAYS-1)
	}

	if this.Hours < 0 {
		return fmt.Errorf("Duration cannot have negative hours, found '%d'.", this.Hours)
	}

	if this.Hours >= MAX_HOURS {
		return fmt.Errorf("Duration has too many hours (%d), max: %d.", this.Hours, MAX_HOURS-1)
	}

	if this.Minutes < 0 {
		return fmt.Errorf("Duration cannot have negative minutes, found '%d'.", this.Minutes)
	}

	if this.Minutes >= MAX_MINS {
		return fmt.Errorf("Duration has too many minutes (%d), max: %d.", this.Minutes, MAX_MINS-1)
	}

	if this.Seconds < 0 {
		return fmt.Errorf("Duration cannot have negative seconds, found '%d'.", this.Seconds)
	}

	if this.Seconds >= MAX_SECS {
		return fmt.Errorf("Duration has too many seconds (%d), max: %d.", this.Seconds, MAX_SECS-1)
	}

	if this.Microseconds < 0 {
		return fmt.Errorf("Duration cannot have negative microseconds, found '%d'.", this.Microseconds)
	}

	if this.Microseconds >= MAX_USECS {
		return fmt.Errorf("Duration has too many microseconds (%d), max: %d.", this.Microseconds, MAX_USECS-1)
	}

	if this.TotalNanosecs() < 0 {
		return fmt.Errorf("Duration is too large and has overflowed.")
	}

	return nil
}

func (this *DurationSpec) TotalNanosecs() int64 {
	return this.Microseconds*int64(time.Microsecond) +
		this.Seconds*int64(time.Second) +
		this.Minutes*int64(time.Minute) +
		this.Hours*int64(time.Hour) +
		this.Days*int64(time.Hour)*24
}

func (this *DurationSpec) IsEmpty() bool {
	return (this.TotalNanosecs() == 0)
}

func (this *DurationSpec) ComputeNextTime(startTime time.Time) time.Time {
	duration := time.Duration(this.TotalNanosecs())
	return startTime.Add(duration)
}

func (this *DurationSpec) String() string {
	return fmt.Sprintf("every %d days, %d hours, %d minutes, %d seconds; (%d total seconds)",
		this.Days, this.Hours, this.Minutes, this.Seconds, (this.TotalNanosecs() / int64(time.Second)))
}

func (this *DurationSpec) ShortString() string {
	parts := make([]string, 0, 4)

	if this.Days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", this.Days))
	}

	if this.Hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hours", this.Hours))
	}

	if this.Minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d minutes", this.Minutes))
	}

	if this.Seconds > 0 {
		parts = append(parts, fmt.Sprintf("%d seconds", this.Seconds))
	}

	return fmt.Sprintf("every %s", strings.Join(parts, ", "))
}

func (this TimeOfDaySpec) Validate() error {
	var contents string = string(this)
	if contents == "" {
		return nil
	}

	_, err := this.getTime()
	return err
}

func (this TimeOfDaySpec) TotalNanosecs() int64 {
	return int64(time.Hour) * 24
}

func (this TimeOfDaySpec) IsEmpty() bool {
	return (string(this) == "")
}

func (this TimeOfDaySpec) ComputeNextTime(startTime time.Time) time.Time {
	var thisTime time.Time

	instance, err := this.getTime()
	if err != nil {
		log.Error("Failed to parse time of day spec.", err, log.NewAttr("contents", string(this)))
		thisTime, _ = time.Parse(TIME_LAYOUT_MINS, "00:00")
	} else {
		thisTime = instance
	}

	// Get a time with the same date as startTime, but the time of day for this scheduled time.
	nextTime := time.Date(
		startTime.Year(), startTime.Month(), startTime.Day(),
		thisTime.Hour(), thisTime.Minute(), thisTime.Second(), thisTime.Nanosecond(),
		startTime.Location())

	// The constructed time may be before the start time.
	for nextTime.Before(startTime) {
		nextTime = nextTime.AddDate(0, 0, 1)
	}

	return nextTime
}

func (this TimeOfDaySpec) String() string {
	timeOfDay := "00:00:00"

	instance, err := this.getTime()
	if err != nil {
		log.Error("Failed to parse time of day spec.", err, log.NewAttr("contents", string(this)))
	} else {
		timeOfDay = instance.Format(time.TimeOnly)
	}

	return fmt.Sprintf("daily at %s.", timeOfDay)
}

// This TimeOfDaySpec should have already been validated,
// so the time should parse.
func (this TimeOfDaySpec) getTime() (time.Time, error) {
	timeOfDay := "00:00:00"

	var contents string = string(this)
	if contents != "" {
		timeOfDay = contents
	}

	timeLayout := ""
	colonCount := strings.Count(timeOfDay, ":")

	if colonCount == 1 {
		timeLayout = TIME_LAYOUT_MINS
	} else if colonCount == 2 {
		timeLayout = TIME_LAYOUT_SECS
	} else {
		return time.Time{}, fmt.Errorf("Time of day does not look like a 24-hour time: '%s'.", timeOfDay)
	}

	instance, err := time.Parse(timeLayout, timeOfDay)
	if err != nil {
		return time.Time{}, fmt.Errorf("Could not parse time of day from '%s', looking for 24-hour time: '%w'.", timeOfDay, err)
	}

	return instance, nil
}

func (this *ScheduledTime) Validate() error {
	err := this.Daily.Validate()
	if err != nil {
		return fmt.Errorf("Schedule time 'daily' component is invalid: '%w'.", err)
	}

	err = this.Every.Validate()
	if err != nil {
		return fmt.Errorf("Schedule time 'every' component is invalid: '%w'.", err)
	}

	if this.Daily.IsEmpty() && this.Every.IsEmpty() {
		return fmt.Errorf("Both 'daily' and 'every' cannot be empty.")
	}

	if !this.Daily.IsEmpty() && !this.Every.IsEmpty() {
		return fmt.Errorf("Both 'daily' and 'every' cannot be populated.")
	}

	return nil
}

func (this *ScheduledTime) TotalNanosecs() int64 {
	if this.Daily.IsEmpty() {
		return this.Every.TotalNanosecs()
	}

	return this.Daily.TotalNanosecs()
}

func (this *ScheduledTime) IsEmpty() bool {
	return (this.Daily.IsEmpty() && this.Every.IsEmpty())
}

func (this *ScheduledTime) ComputeNextTimeFromNow() time.Time {
	return this.ComputeNextTime(time.Now())
}

func (this *ScheduledTime) ComputeNextTime(startTime time.Time) time.Time {
	if this.Daily.IsEmpty() {
		return this.Every.ComputeNextTime(startTime)
	}

	return this.Daily.ComputeNextTime(startTime)
}

func (this *ScheduledTime) String() string {
	if this.Daily.IsEmpty() {
		return this.Every.String()
	}

	return this.Daily.String()
}
