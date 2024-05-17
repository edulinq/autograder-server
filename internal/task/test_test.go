package task

import (
    "testing"
    "time"

    "github.com/edulinq/autograder/internal/common"
    "github.com/edulinq/autograder/internal/config"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/model/tasks"
)

func TestTaskBase(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    // Set the task rest time to negative.
    oldRestTime := config.TASK_MIN_REST_SECS.Get();
    config.TASK_MIN_REST_SECS.Set(-1);
    defer config.TASK_MIN_REST_SECS.Set(oldRestTime);

    count := runTestTask(test, 5);

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

    count := runTestTask(test, 5);

    if (count != 1) {
        test.Fatalf("Incorrect number of runs. Expected exactly 1, got %d.", count);
    }
}

// Ensure that a catchup task will run.
func TestTaskCatchup(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    // Set the task rest time to negative.
    oldRestTime := config.TASK_MIN_REST_SECS.Get();
    config.TASK_MIN_REST_SECS.Set(-1);
    defer config.TASK_MIN_REST_SECS.Set(oldRestTime);

    // Set the last run for this task to be far in the past
    // (but not a zero time).
    err := db.LogTaskCompletion("course101", "course101::test", time.Time{}.Add(time.Second));
    if (err != nil) {
        test.Fatalf("Failed to log task completion: '%v'.", err);
    }

    // Set the duration high enough so it will never run.
    count := runTestTask(test, 100000000);

    // Exactly one instance of the task (the catchup) should have run.
    if (count != 1) {
        test.Fatalf("Incorrect number of runs. Expected exactly 1, got %d.", count);
    }
}

// Ensure that a catchup task will not run.
func TestTaskNoCatchup(test *testing.T) {
    db.ResetForTesting();
    defer db.ResetForTesting();

    // Set the task rest time to negative.
    oldRestTime := config.TASK_MIN_REST_SECS.Get();
    config.TASK_MIN_REST_SECS.Set(-1);
    defer config.TASK_MIN_REST_SECS.Set(oldRestTime);

    // Set the last run for this task to be right now.
    err := db.LogTaskCompletion("course101", "course101::test", time.Now());
    if (err != nil) {
        test.Fatalf("Failed to log task completion: '%v'.", err);
    }

    // Set the duration high enough so it will never run.
    count := runTestTask(test, 100000000);

    // No tasks should run.
    if (count != 0) {
        test.Fatalf("Incorrect number of runs. Expected exactly 0, got %d.", count);
    }
}

// Test that panics are recovered.
func TestPanic(test *testing.T) {
    counter := make(chan int, 100);
    fun := func(payload any) error {
        counter := payload.(chan int);
        counter <- 1;
        panic("Test Panic");
        return nil;
    }

    count := runTestTaskWithFun(test, 1, counter, fun);
    if (count <= 1) {
        test.Fatalf("Not enough test tasks were run (%d run).", count)
    }
}

// Run a basic test task.
// Return the number of times the task was run.
func runTestTask(test *testing.T, everyUSecs int64) int {
    counter := make(chan int, 100);

    fun := func(payload any) error {
        counter := payload.(chan int);
        counter <- 1;
        return nil;
    }

    return runTestTaskWithFun(test, everyUSecs, counter, fun);
}

func runTestTaskWithFun(test *testing.T, everyUSecs int64, counter chan int, fun func(any) error) int {
    defer StopAll();

    course := db.MustGetTestCourse();

    task := &tasks.TestTask{
        BaseTask: &tasks.BaseTask{
            Disable: false,
            When: []*common.ScheduledTime{
                &common.ScheduledTime{
                    Every: common.DurationSpec{
                        Microseconds: everyUSecs,
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
    time.Sleep(1750 * time.Microsecond);

    // Stop the task.
    StopCourse(course.GetID());
    close(counter);

    count := 0;
    for _ = range counter {
        count++;
    }

    return count;
}
