package task

import (
    "testing"
    "time"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func TestTaskBase(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    // Set the task rest time to negative.
    oldRestTime := config.TASK_MIN_REST_SECS.Get();
    config.TASK_MIN_REST_SECS.Set(-1);
    defer config.TASK_MIN_REST_SECS.Set(oldRestTime);

    count := runTestTask(test);

    if (count <= 1) {
        test.Fatalf("Not enough test tasks were run (%d run). (It's possible for this to be flaky on a very busy machine).", count)
    }
}

// Test that tasks will not run too often.
func TestTaskSkipRecent(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    // Set the task rest time to something large.
    oldRestTime := config.TASK_MIN_REST_SECS.Get();
    config.TASK_MIN_REST_SECS.Set(10 * 60);
    defer config.TASK_MIN_REST_SECS.Set(oldRestTime);

    count := runTestTask(test);

    if (count != 1) {
        test.Fatalf("Incorrect number of runs. Expected exactly 1, got %d.", count);
    }
}

// Run a basic test task.
// Return the number of times the task was run.
func runTestTask(test *testing.T) int {
    defer StopAll();

    counter := make(chan int, 100);

    fun := func(payload any) error {
        counter := payload.(chan int);
        counter <- 1;
        return nil;
    }

    course := db.MustGetTestCourse();

    task := &tasks.TestTask{
        BaseTask: &tasks.BaseTask{
            Disable: false,
            When: []*tasks.ScheduledTime{
                &tasks.ScheduledTime{
                    Every: tasks.DurationSpec{
                        Microseconds: 5,
                    },
                },
            },
        },
        Func: fun,
        Payload: counter,
    };

    err := task.Validate(course);
    if (err != nil) {
        test.Fatalf("Failed to validate test course: '%v'.", err);
    }

    // Start the task.
    err = Schedule(course, task);
    if (err != nil) {
        test.Fatalf("Failed to schedule task: '%v'.", err);
    }

    // Wait.
    time.Sleep(750 * time.Microsecond);

    // Stop the task.
    StopCourse(course.GetID());
    close(counter);

    count := 0;
    for _ = range counter {
        count++;
    }

    return count;
}
