package tasks

import (
	"testing"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

func TestRunCourseScoringUploadTaskBase(test *testing.T) {
	task := &model.FullScheduledTask{
		SystemTaskInfo: model.SystemTaskInfo{
			CourseID: db.TEST_COURSE_ID,
		},
	}

	err := RunCourseScoringUploadTask(task)
	if err != nil {
		test.Fatalf("Got an unexpected error running task: '%v'.", err)
	}
}
