package tasks

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/logs"
)

func RunCourseEmailLogsTask(task *model.FullScheduledTask) error {
	course, err := db.GetCourse(task.CourseID)
	if err != nil {
		return fmt.Errorf("Failed to get course '%s': '%w'.", task.CourseID, err)
	}

	if course == nil {
		return fmt.Errorf("Unable to find course '%s'.", task.CourseID)
	}

	to, err := model.GetTaskOptionAsType(&task.UserTaskInfo, "to", []string{})
	if err != nil {
		return fmt.Errorf("Unable to get recipients: '%w'.", err)
	}

	rawQuery, err := model.GetTaskOptionAsType(&task.UserTaskInfo, "query", log.RawLogQuery{})
	if err != nil {
		return fmt.Errorf("Unable to get query: '%w'.", err)
	}

	sendEmpty := (task.Options["send-empty"] == true)

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

	return nil
}
