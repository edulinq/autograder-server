package task

import (
    "fmt"
    "strings"

    "github.com/edulinq/autograder/internal/common"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/email"
    "github.com/edulinq/autograder/internal/log"
    "github.com/edulinq/autograder/internal/model"
    "github.com/edulinq/autograder/internal/model/tasks"
)

func RunEmailLogsTask(course *model.Course, rawTask tasks.ScheduledTask) (bool, error) {
    task, ok := rawTask.(*tasks.EmailLogsTask);
    if (!ok) {
        return false, fmt.Errorf("Task is not a EmailLogsTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return true, nil;
    }

    return true, RunEmailLogs(task.RawLogQuery, course, task.To, task.SendEmpty);
}

func RunEmailLogs(rawQuery common.RawLogQuery, course *model.Course, to []string, sendEmpty bool) error {
    parsedQuery, err := rawQuery.ParseJoin(course);
    if (err != nil) {
        return err;
    }

    if (parsedQuery.UserID != "") {
        fullUser, err := db.GetUser(course, parsedQuery.UserID);
        if (err != nil) {
            return err;
        }

        if (fullUser == nil) {
            return fmt.Errorf("Could not find user: '%s'.", parsedQuery.UserID);
        } else {
            parsedQuery.UserID = fullUser.Email;
        }
    }

    records, err := db.GetLogRecords(parsedQuery.Level, parsedQuery.After,
            course.GetID(), parsedQuery.AssignmentID, parsedQuery.UserID);
    if (err != nil) {
        return fmt.Errorf("Failed to get log records: '%v'.", err);
    }

    var content strings.Builder;
    content.WriteString(fmt.Sprintf("Found %d log records matching query: [%s].\n", len(records), parsedQuery.String()));

    if ((len(records) == 0) && !sendEmpty) {
        return nil;
    }

    for _, record := range records {
        content.WriteString("\n");
        content.WriteString(record.String());
    }

    subject := fmt.Sprintf("Autograder Logs for %s", course.GetName());

    err = email.Send(to, subject, content.String(), false);
    if (err != nil) {
        return fmt.Errorf("Failed to send logs for course '%s': '%w'.", course.GetName(), err);
    }

    log.Debug("EmailLogs completed sucessfully.", course, log.NewAttr("to", to));
    return nil;
}
