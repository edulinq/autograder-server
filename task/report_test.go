package task

import (
    "reflect"
    "strings"
    "testing"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func TestReportBase(test *testing.T) {
    db.ResetForTesting();

    course := db.MustGetTestCourse();

    to := []string{"test1@test.com", "test2@test.com"};

    task := &tasks.ReportTask{
        BaseTask: &tasks.BaseTask{
            Disable: false,
            When: []*tasks.ScheduledTime{},
        },
        To: to,
    };

    err := RunReportTask(course, task);
    if (err != nil) {
        test.Fatalf("Failed to run report task: '%v'.", err);
    }

    messages := email.GetTestMessages();

    if (len(messages) != 1) {
        test.Fatalf("Did not find the correct number of messages. Expected: 1, Found: %d.", len(messages));
    }

    if (!reflect.DeepEqual(to, messages[0].To)) {
        test.Fatalf("Unexpected message recipients. Expected: [%s], Found: [%s].",
            strings.Join(to, ", "), strings.Join(messages[0].To, ", "));
    }
}
