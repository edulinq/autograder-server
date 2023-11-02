package task

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/model2"
    "github.com/eriq-augustine/autograder/report"
)

type ReportTask struct {
    Disable bool `json:"disable"`
    When ScheduledTime `json:"when"`

    To []string `json:"to"`

    course model2.Course `json:"-"`
}

func (this *ReportTask) Validate(course model2.Course) error {
    this.When.id = fmt.Sprintf("report-%s", course.GetID());

    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.Get());
    this.course = course;

    if (!this.Disable && (len(this.To) == 0)) {
        return fmt.Errorf("Report task is not disabled, but no email recipients are declared.");
    }

    return nil;
}

func (this *ReportTask) String() string {
    return fmt.Sprintf("Report on '%s' at '%s' (next time: '%s').", this.course.GetName(), this.When.String(), this.When.ComputeNext());
}

// Schedule this task to be regularly run at the scheduled time.
func (this *ReportTask) Schedule() {
    if (this.Disable) {
        return;
    }

    this.When.Schedule(func() {
        err := this.Run();
        if (err != nil) {
            log.Error().Err(err).Str("course", this.course.GetName()).Msg("Report task failed.");
        }
    });
}

// Stop any scheduled executions of this task.
func (this *ReportTask) Stop() {
    this.When.Stop();
}

// Run the task regardless of schedule.
func (this *ReportTask) Run() error {
    return RunReport(this.course, this.To);
}

// Do a report without an attatched object.
func RunReport(course model2.Course, to []string) error {
    report, err := report.GetCourseScoringReport(course);
    if (err != nil) {
        return fmt.Errorf("Failed to get scoring report for course '%s': '%w'.", course.GetName(), err);
    }

    html, err := report.ToHTML();
    if (err != nil) {
        return fmt.Errorf("Failed to generate HTML for scoring report for course '%s': '%w'.", course.GetName(), err);
    }

    subject := fmt.Sprintf("Autograder Scoring Report for %s", course.GetName());

    err = email.Send(to, subject, html, true);
    if (err != nil) {
        return fmt.Errorf("Failed to send scoring report for course '%s': '%w'.", course.GetName(), err);
    }

    log.Debug().Str("course", course.GetName()).Any("to", to).Msg("Report completed sucessfully.");
    return nil;
}
