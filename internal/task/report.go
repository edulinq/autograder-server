package task

import (
    "fmt"

    "github.com/edulinq/autograder/internal/email"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/model"
    "github.com/edulinq/autograder/internal/model/tasks"
    "github.com/edulinq/autograder/internal/report"
)

func RunReportTask(course *model.Course, rawTask tasks.ScheduledTask) (bool, error) {
    task, ok := rawTask.(*tasks.ReportTask);
    if (!ok) {
        return false, fmt.Errorf("Task is not a ReportTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return true, nil;
    }

    return true, RunReport(course, task.To);
}

func RunReport(course *model.Course, to []string) error {
    report, err := report.GetCourseScoringReport(course);
    if (err != nil) {
        return fmt.Errorf("Failed to get scoring report for course '%s': '%w'.", course.GetID(), err);
    }

    html, err := report.ToHTML();
    if (err != nil) {
        return fmt.Errorf("Failed to generate HTML for scoring report for course '%s': '%w'.", course.GetID(), err);
    }

    subject := fmt.Sprintf("Autograder Scoring Report for %s", course.GetName());

    to, err = db.ResolveUsers(course, to);
    if (err != nil) {
        return fmt.Errorf("Failed to resolve users for course '%s': '%w'.", course.GetID(), err);
    }

    err = email.Send(to, subject, html, true);
    if (err != nil) {
        return fmt.Errorf("Failed to send scoring report for course '%s': '%w'.", course.GetID(), err);
    }

    log.Debug("Report completed sucessfully.", course, log.NewAttr("to", to));
    return nil;
}
