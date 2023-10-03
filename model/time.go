package model

import (
    "fmt"
    "time"
    "strings"
)

const (
    TIME_LAYOUT_MINS = "15:04";
    TIME_LAYOUT_SECS = "15:04:05";

    DEFAULT_TIME_OF_DAY = "00:00";
)

var dayOfWeekStrings map[string]time.Weekday = map[string]time.Weekday{
    "sun": time.Sunday,
    "mon": time.Monday,
    "tue": time.Tuesday,
    "wed": time.Wednesday,
    "thu": time.Thursday,
    "fri": time.Friday,
    "sat": time.Saturday,
};

// This struct should always have Validate() called after construction.
// All other methods will assume Validate() returns no error.
type ScheduledTime struct {
    DayOfWeek string `json:"day-of-week"`
    TimeOfDay string `json:"time-of-day"`

    timeLayout string
    nextRun time.Time
    timer *time.Timer
}

func NewScheduledTime(dayOfWeek string, timeOfDay string) ScheduledTime {
    return ScheduledTime{
        DayOfWeek: dayOfWeek,
        TimeOfDay: timeOfDay,
        nextRun: time.Time{},
        timer: nil,
    };
}

func (this *ScheduledTime) Validate() error {
    if (this.TimeOfDay == "") {
        this.TimeOfDay = DEFAULT_TIME_OF_DAY;
    }

    colonCount := strings.Count(this.TimeOfDay, ":");

    if (colonCount == 1) {
        this.timeLayout = TIME_LAYOUT_MINS;
    } else if (colonCount == 2) {
        this.timeLayout = TIME_LAYOUT_SECS;
    } else {
        return fmt.Errorf("Time of day does not look like a 24-hour time: '%s'.", this.TimeOfDay);
    }

    _, err := time.Parse(this.timeLayout, this.TimeOfDay);
    if (err != nil) {
        return fmt.Errorf("Could not parse time of day from '%s', looking for 24-hour time: '%w'.", this.TimeOfDay, err);
    }

    if (this.DayOfWeek != "") {
        if (len(this.DayOfWeek) < 3) {
            return fmt.Errorf("Unknown day of week '%s'. Must be full day name in English, first three letters, or nothing (everyday).", this.DayOfWeek);
        }

        rawDayOfWeek := this.DayOfWeek;
        this.DayOfWeek = strings.ToLower(this.DayOfWeek)[0:3]

        _, ok := dayOfWeekStrings[this.DayOfWeek];
        if (!ok) {
            return fmt.Errorf("Unknown day of week '%s' (parsed to '%s'). Must be full day name in English, first three letters, or nothing (everyday).", rawDayOfWeek, this.DayOfWeek);
        }
    }

    return nil;
}

func (this *ScheduledTime) String() string {
    dayOfWeek := this.DayOfWeek;
    if (dayOfWeek == "") {
        dayOfWeek = "everyday";
    }

    return fmt.Sprintf("%s at %s", dayOfWeek, this.TimeOfDay);
}

// Compute when the next time this scheduled time will occur,
// but do not actually schedule or run anything.
func (this *ScheduledTime) ComputeNext() time.Time {
    return this.computeNextTime(time.Now());
}

// Compute the next time to run starting at the passed in time.
// Note that the passed in time can also be a valid time to run.
func (this *ScheduledTime) computeNextTime(startTime time.Time) time.Time {
    parsedTimeOfDay := this.GetTimeOfDay();

    // Get a time with the same date as startTime, but the time of day for this scheduled time.
    nextTime := time.Date(
            startTime.Year(), startTime.Month(), startTime.Day(),
            parsedTimeOfDay.Hour(), parsedTimeOfDay.Minute(), parsedTimeOfDay.Second(), parsedTimeOfDay.Nanosecond(),
            startTime.Location());

    for (nextTime.Before(startTime) || !this.isCorrectDayOfWeek(nextTime)) {
        nextTime = nextTime.AddDate(0, 0, 1);
    }

    return nextTime;
}

// Set a recurring invoation of the given task.
func (this *ScheduledTime) Schedule(task func()) {
    // Cleanup any old timers.
    this.Stop();

    this.nextRun = this.ComputeNext();
    this.timer = time.AfterFunc(this.nextRun.Sub(time.Now()), func() {
        task()
        this.Schedule(task);
    });
}

func (this *ScheduledTime) Stop() {
    if (this.timer == nil) {
        return;
    }

    this.nextRun = time.Time{}

    if (!this.timer.Stop()) {
        // Clear the channel.
        <- this.timer.C;
    }
    this.timer = nil;
}

func (this *ScheduledTime) GetNextRunTime() time.Time {
    return this.nextRun;
}

func (this *ScheduledTime) GetTimeOfDay() time.Time {
    timeOfDay, _ := time.Parse(this.timeLayout, this.TimeOfDay);
    return timeOfDay;
}

func (this *ScheduledTime) isCorrectDayOfWeek(someTime time.Time) bool {
    if (this.DayOfWeek == "") {
        return true;
    }

    return (someTime.Weekday() == dayOfWeekStrings[this.DayOfWeek]);
}
