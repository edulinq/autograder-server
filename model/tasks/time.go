package tasks

import (
    "fmt"
    "time"
    "strings"

    "github.com/rs/zerolog/log"
)

const (
    TIME_LAYOUT_MINS = "15:04";
    TIME_LAYOUT_SECS = "15:04:05";
)

// This struct should always have Validate() called after construction.
// All other methods will assume Validate() returns no error.
type ScheduledTime struct {
    Every DurationSpec `json:"every"`
    Daily TimeOfDaySpec `json:"daily"`
}

type timeSpec interface {
    Validate() error
    TotalNanosecs() int64
    IsEmpty() bool
    // Get the next time to run (starting at the passed in time).
    ComputeNextTime(startTime time.Time) time.Time
    String() string
}

type TimeOfDaySpec string;

type DurationSpec struct {
    Days int64 `json:"days"`
    Hours int64 `json:"hours"`
    Minutes int64 `json:"minutes"`
    Seconds int64 `json:"seconds"`

    // Microseconds in not exposed in JSON and in meant for testing.
    Microseconds int64 `json:"-"`
}

func (this *DurationSpec) Validate() error {
    if (this.Days < 0) {
        return fmt.Errorf("Duration cannot have negative days, found '%d'.", this.Days);
    }

    if (this.Hours < 0) {
        return fmt.Errorf("Duration cannot have negative hours, found '%d'.", this.Hours);
    }

    if (this.Minutes < 0) {
        return fmt.Errorf("Duration cannot have negative minutes, found '%d'.", this.Minutes);
    }

    if (this.Seconds < 0) {
        return fmt.Errorf("Duration cannot have negative seconds, found '%d'.", this.Seconds);
    }

    if (this.Microseconds < 0) {
        return fmt.Errorf("Duration cannot have negative microseconds, found '%d'.", this.Microseconds);
    }

    return nil;
}

func (this *DurationSpec) TotalNanosecs() int64 {
    return this.Microseconds * int64(time.Microsecond) +
        this.Seconds * int64(time.Second) +
        this.Minutes * int64(time.Minute) +
        this.Hours * int64(time.Hour) +
        this.Days * int64(time.Hour) * 24;
}

func (this *DurationSpec) IsEmpty() bool {
    return (this.TotalNanosecs() == 0);
}

func (this *DurationSpec) ComputeNextTime(startTime time.Time) time.Time {
    duration := time.Duration(this.TotalNanosecs());
    return startTime.Add(duration);
}

func (this *DurationSpec) String() string {
    return fmt.Sprintf("Every %d days, %d hours, %d minutes, %d seconds; (%d total seconds)",
        this.Days, this.Hours, this.Minutes, this.Seconds, (this.TotalNanosecs() / int64(time.Second)));
}

func (this TimeOfDaySpec) Validate() error {
    var contents string = string(this);
    if (contents == "") {
        return nil;
    }

    _, err := this.getTime();
    return err;
}

func (this TimeOfDaySpec) TotalNanosecs() int64 {
    return int64(time.Hour) * 24;
}

func (this TimeOfDaySpec) IsEmpty() bool {
    return (string(this) == "");
}

func (this TimeOfDaySpec) ComputeNextTime(startTime time.Time) time.Time {
    var thisTime time.Time;

    instance, err := this.getTime();
    if (err != nil) {
        log.Error().Err(err).Str("contents", string(this)).Msg("Failed to parse time of day spec.");
        thisTime, _ = time.Parse(TIME_LAYOUT_MINS, "00:00");
    } else {
        thisTime = instance;
    }

    // Get a time with the same date as startTime, but the time of day for this scheduled time.
    nextTime := time.Date(
            startTime.Year(), startTime.Month(), startTime.Day(),
            thisTime.Hour(), thisTime.Minute(), thisTime.Second(), thisTime.Nanosecond(),
            startTime.Location());

    // The constructed time may be before the start time.
    for (nextTime.Before(startTime)) {
        nextTime = nextTime.AddDate(0, 0, 1);
    }

    return nextTime;
}

func (this TimeOfDaySpec) String() string {
    timeOfDay := "00:00:00";

    instance, err := this.getTime();
    if (err != nil) {
        log.Error().Err(err).Str("contents", string(this)).Msg("Failed to parse time of day spec.");
    } else {
        timeOfDay = instance.Format(time.TimeOnly);
    }

    return fmt.Sprintf("Daily at %s.", timeOfDay);
}

// This TimeOfDaySpec should have already been validated,
// so the time should parse.
func (this TimeOfDaySpec) getTime() (time.Time, error) {
    var contents string = string(this);
    if (contents == "") {
        return time.Time{}, fmt.Errorf("Time of day spec is empty, cannot parse.");
    }

    timeLayout := "";
    colonCount := strings.Count(contents, ":");

    if (colonCount == 1) {
        timeLayout = TIME_LAYOUT_MINS;
    } else if (colonCount == 2) {
        timeLayout = TIME_LAYOUT_SECS;
    } else {
        return time.Time{}, fmt.Errorf("Time of day does not look like a 24-hour time: '%s'.", contents);
    }

    instance, err := time.Parse(timeLayout, contents);
    if (err != nil) {
        return time.Time{}, fmt.Errorf("Could not parse time of day from '%s', looking for 24-hour time: '%w'.", contents, err);
    }

    return instance, nil;
}

func (this *ScheduledTime) Validate() error {
    err := this.Daily.Validate();
    if (err != nil) {
        return fmt.Errorf("Schedule time 'daily' component is invalid: '%w'.", err);
    }

    err = this.Every.Validate();
    if (err != nil) {
        return fmt.Errorf("Schedule time 'every' component is invalid: '%w'.", err);
    }

    if (this.Daily.IsEmpty() && this.Every.IsEmpty()) {
        return fmt.Errorf("Both 'daily' and 'every' cannot be empty.");
    }

    if (!this.Daily.IsEmpty() && !this.Every.IsEmpty()) {
        return fmt.Errorf("Both 'daily' and 'every' cannot be populated.");
    }

    return nil;
}

func (this *ScheduledTime) TotalNanosecs() int64 {
    if (this.Daily.IsEmpty()) {
        return this.Every.TotalNanosecs();
    }

    return this.Daily.TotalNanosecs();
}

func (this *ScheduledTime) IsEmpty() bool {
    return (this.Daily.IsEmpty() && this.Every.IsEmpty());
}

func (this *ScheduledTime) ComputeNextTimeFromNow() time.Time {
    return this.ComputeNextTime(time.Now());
}

func (this *ScheduledTime) ComputeNextTime(startTime time.Time) time.Time {
    if (this.Daily.IsEmpty()) {
        return this.Every.ComputeNextTime(startTime);
    }

    return this.Daily.ComputeNextTime(startTime);
}

func (this *ScheduledTime) String() string {
    if (this.Daily.IsEmpty()) {
        return this.Every.String();
    }

    return this.Daily.String();
}
