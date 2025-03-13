package tasks

import (
	"github.com/edulinq/autograder/internal/model"
)

var testTaskCalls int = 0

func resetTestTaskCalls() {
	testTaskCalls = 0
}

func RunTestTask(task *model.FullScheduledTask) error {
	testTaskCalls++
	return nil
}
