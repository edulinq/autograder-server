package tasks

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/report"
)

func RunCourseReportTask(task *model.FullScheduledTask) error {
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

	report, err := report.GetCourseScoringReport(course)
	if err != nil {
		return fmt.Errorf("Failed to get scoring report for course '%s': '%w'.", course.GetID(), err)
	}

	html, err := report.ToHTML()
	if err != nil {
		return fmt.Errorf("Failed to generate HTML for scoring report for course '%s': '%w'.", course.GetID(), err)
	}

	subject := fmt.Sprintf("Autograder Scoring Report for %s", course.GetName())

	to, err = db.ResolveCourseUsers(course, to)
	if err != nil {
		return fmt.Errorf("Failed to resolve users for course '%s': '%w'.", course.GetID(), err)
	}

	err = email.Send(to, subject, html, true)
	if err != nil {
		return fmt.Errorf("Failed to send scoring report for course '%s': '%w'.", course.GetID(), err)
	}

	log.Debug("Report completed successfully.", course, log.NewAttr("to", to))
	return nil
}
