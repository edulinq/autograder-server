package tasks

import (
	"testing"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

func TestTaskCoreRunOneTask(test *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	resetTestTaskCalls()
	defer resetTestTaskCalls()

	enableTaskEngine = true
	defer func() {
		enableTaskEngine = false
	}()

	task := &model.FullScheduledTask{
		UserTaskInfo: model.UserTaskInfo{
			Type: model.TaskTypeTest,
			When: &common.ScheduledTime{
				Daily: "0:00",
			},
		},
		SystemTaskInfo: model.SystemTaskInfo{
			Source:      model.TaskSourceTest,
			LastRunTime: timestamp.Zero(),
			NextRunTime: timestamp.Zero(),
			Hash:        "ABC",
		},
	}

	db.MustUpsertActiveTask(task)

	if testTaskCalls != 0 {
		test.Fatalf("Intial value for test task is wrong. Expected: %d, Actual: %d.", 0, testTaskCalls)
	}

	runNextTask()

	if testTaskCalls != 1 {
		test.Fatalf("Final value for test task is wrong. Expected: %d, Actual: %d.", 1, testTaskCalls)
	}
}
