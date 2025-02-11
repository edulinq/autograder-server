package tasks

import (
	"testing"

	"github.com/edulinq/autograder/internal/model"
)

func TestRunTestTaskBase(test *testing.T) {
	resetTestTaskCalls()
	defer resetTestTaskCalls()

	task := &model.FullScheduledTask{}

	err := RunTestTask(task)
	if err != nil {
		test.Fatalf("Got an unexpected error running task: '%v'.", err)
	}

	if testTaskCalls != 1 {
		test.Fatalf("Test task did not properly invoke function.")
	}
}
