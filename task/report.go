package task

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/report"
)

type ReportTask struct {
    Disable bool `json:"disable"`
    When ScheduledTime `json:"when"`

    To []string `json:"to"`

    sources report.ReportingSources `json:"-"`
}

func (this *ReportTask) Validate(course TaskCourseSource) error {
    this.When.id = fmt.Sprintf("report-%s", course.GetID());

    err := this.When.Validate();
    if (err != nil) {
        return err;
    }

    this.Disable = (this.Disable || config.NO_TASKS.Get());
    this.sources = course;

    if (!this.Disable && (len(this.To) == 0)) {
        return fmt.Errorf("Report task is not disabled, but no email recipients are declared.");
    }

    return nil;
}

func (this *ReportTask) String() string {
    return fmt.Sprintf("Report on '%s' at '%s' (next time: '%s').", this.sources.GetName(), this.When.String(), this.When.ComputeNext());
}

// Schedule this task to be regularly run at the scheduled time.
func (this *ReportTask) Schedule() {
    if (this.Disable) {
        return;
    }

    this.When.Schedule(func() {
        err := this.Run();
        if (err != nil) {
            log.Error().Err(err).Str("course", this.sources.GetName()).Msg("Report task failed.");
        }
    });
}

// Stop any scheduled executions of this task.
func (this *ReportTask) Stop() {
    this.When.Stop();
}

// Run the task regardless of schedule.
func (this *ReportTask) Run() error {
    return RunReport(this.sources, this.To);
}

// Do a report without an attatched object.
func RunReport(sources report.ReportingSources, to []string) error {
    report, err := report.GetCourseScoringReport(sources);
    if (err != nil) {
        return fmt.Errorf("Failed to get scoring report for course '%s': '%w'.", sources.GetName(), err);
    }

    html, err := report.ToHTML();
    if (err != nil) {
        return fmt.Errorf("Failed to generate HTML for scoring report for course '%s': '%w'.", sources.GetName(), err);
    }

    subject := fmt.Sprintf("Autograder Scoring Report for %s", sources.GetName());

    err = email.Send(to, subject, html, true);
    if (err != nil) {
        return fmt.Errorf("Failed to send scoring report for course '%s': '%w'.", sources.GetName(), err);
    }

    log.Debug().Str("course", sources.GetName()).Any("to", to).Msg("Report completed sucessfully.");
    return nil;
}
