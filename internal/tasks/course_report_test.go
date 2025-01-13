package tasks

import (
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

func TestRunCourseReportTaskBase(test *testing.T) {
	task := &model.FullScheduledTask{
		UserTaskInfo: model.UserTaskInfo{
			Options: map[string]any{
				"to": []string{"course-admin@test.edulinq.org"},
			},
		},
		SystemTaskInfo: model.SystemTaskInfo{
			CourseID: db.TEST_COURSE_ID,
		},
	}

	err := RunCourseReportTask(task)
	if err != nil {
		test.Fatalf("Got an unexpected error on backup: '%v'.", err)
	}
}
