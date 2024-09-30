package task

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/model/tasks"
	"github.com/edulinq/autograder/internal/procedures/logs"
)

func RunEmailLogsTask(course *model.Course, rawTask tasks.ScheduledTask) (bool, error) {
	task, ok := rawTask.(*tasks.EmailLogsTask)
	if !ok {
		return false, fmt.Errorf("Task is not a EmailLogsTask: %t (%v).", rawTask, rawTask)
	}

	if task.Disable {
		return true, nil
	}

	return true, RunEmailLogs(task.RawLogQuery, course, task.To, task.SendEmpty)
}

func RunEmailLogs(rawQuery log.RawLogQuery, course *model.Course, to []string, sendEmpty bool) error {
	// This query can only be for the context course.
	rawQuery.CourseID = course.ID

	// We will query the logs using root.
	// Since we have already limited the query context to this course, this is safe.
	user, err := db.GetServerUser(model.RootUserEmail)
	if err != nil {
		return fmt.Errorf("Could not get root user: '%w'.", err)
	}

	if user == nil {
		return fmt.Errorf("Could not find root user.")
	}

	records, locatableErr, err := logs.Query(rawQuery, user)
	if err != nil {
		return err
	}

	if locatableErr != nil {
		return locatableErr.ToError()
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("Found %d log records matching query: [%s].\n", len(records), rawQuery.String()))

	if (len(records) == 0) && !sendEmpty {
		return nil
	}

	for _, record := range records {
		content.WriteString("\n")
		content.WriteString(record.String())
	}

	subject := fmt.Sprintf("Autograder Logs for %s", course.GetName())

	err = email.Send(to, subject, content.String(), false)
	if err != nil {
		return fmt.Errorf("Failed to send logs for course '%s': '%w'.", course.GetName(), err)
	}

	log.Debug("EmailLogs completed sucessfully.", course, log.NewAttr("to", to))
	return nil
}
