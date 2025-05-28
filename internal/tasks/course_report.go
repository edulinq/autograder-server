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

	to, err := model.GetTaskOptionAsType(&task.UserTaskInfo, "to", []model.CourseUserReference{})
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

	users, err := db.GetCourseUsers(course)
	if err != nil {
		return fmt.Errorf("Failed to get course users for course '%s': '%w'.", course.GetID(), err)
	}

	reference, userErrors := model.ParseCourseUserReferences(to)
	if userErrors != nil {
		return fmt.Errorf("Failed to parse course user references: '%v'.", userErrors)
	}

	emailTo := model.ResolveCourseUserEmails(users, reference)

	err = email.Send(emailTo, subject, html, true)
	if err != nil {
		return fmt.Errorf("Failed to send scoring report for course '%s': '%w'.", course.GetID(), err)
	}

	log.Debug("Report completed successfully.", course, log.NewAttr("to", to))
	return nil
}
