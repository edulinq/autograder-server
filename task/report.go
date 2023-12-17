package task

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/model/tasks"
    "github.com/eriq-augustine/autograder/report"
)

func RunReportTask(course *model.Course, rawTask tasks.ScheduledTask) error {
    task, ok := rawTask.(*tasks.ReportTask);
    if (!ok) {
        return fmt.Errorf("Task is not a ReportTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return nil;
    }

    return RunReport(course, task.To);
}

func RunReport(course *model.Course, to []string) error {
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
