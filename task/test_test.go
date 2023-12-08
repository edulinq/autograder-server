package task

import (
    "testing"
    "time"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func TestTaskBase(test *testing.T) {
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

    if (count == 0) {
        test.Fatalf("No test tasks were run. (It's possible for this to be flaky on a very busy machine).")
    }
}
