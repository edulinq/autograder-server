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

    SECS_PER_MIN = 60;
    SECS_PER_HOUR = SECS_PER_MIN * 60;
    SECS_PER_DAY = SECS_PER_HOUR * 24;
)

// This struct should always have Validate() called after construction.
// All other methods will assume Validate() returns no error.
type ScheduledTime struct {
    Every durationSpec `json:"every"`
    Daily timeOfDaySpec `json:"daily"`
}

type timeSpec interface {
    Validate() error
    TotalSecs() int
    IsEmpty() bool
    // Get the next time to run (starting at the passed in time).
    ComputeNextTime(startTime time.Time) time.Time
    String() string
}

type timeOfDaySpec string;

type durationSpec struct {
    Days int `json:"days"`
    Hours int `json:"hours"`
    Minutes int `json:"minutes"`
    Seconds int `json:"seconds"`
}

func (this *durationSpec) Validate() error {
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

    return nil;
}

func (this *durationSpec) TotalSecs() int {
    return this.Seconds +
        this.Minutes * 60 +
        this.Hours * 60 * 60 +
        this.Days * 24 * 60 * 60;
}

func (this *durationSpec) IsEmpty() bool {
    return (this.TotalSecs() == 0);
}

func (this *durationSpec) ComputeNextTime(startTime time.Time) time.Time {
    duration := time.Duration(int64(this.TotalSecs()) * int64(time.Second));
    return startTime.Add(duration);
}

func (this *durationSpec) String() string {
    return fmt.Sprintf("Every %d days, %d hours, %d minutes, %d seconds; (%d total seconds)",
        this.Days, this.Hours, this.Minutes, this.Seconds, this.TotalSecs());
}

func (this timeOfDaySpec) Validate() error {
    var contents string = string(this);
    if (contents == "") {
        return nil;
    }

    _, err := this.getTime();
    return err;
}

func (this timeOfDaySpec) TotalSecs() int {
    return 24 * 60 * 60;
}

func (this timeOfDaySpec) IsEmpty() bool {
    return (string(this) == "");
}

func (this timeOfDaySpec) ComputeNextTime(startTime time.Time) time.Time {
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

func (this timeOfDaySpec) String() string {
    timeOfDay := "00:00:00";

    instance, err := this.getTime();
    if (err != nil) {
        log.Error().Err(err).Str("contents", string(this)).Msg("Failed to parse time of day spec.");
    } else {
        timeOfDay = instance.Format(time.TimeOnly);
    }

    return fmt.Sprintf("Daily at %s.", timeOfDay);
}

// This timeOfDaySpec should have already been validated,
// so the time should parse.
func (this timeOfDaySpec) getTime() (time.Time, error) {
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

func (this *ScheduledTime) TotalSecs() int {
    if (this.Daily.IsEmpty()) {
        return this.Every.TotalSecs();
    }

    return this.Daily.TotalSecs();
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
