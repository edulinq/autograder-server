package db

import (
	"reflect"
	"slices"
	"testing"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestUpsertActiveTasksBase(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	tasks, err := GetActiveTasks()
	if err != nil {
		test.Fatalf("Failed to fetch empty tasks: '%v'.", err)
	}

	if len(tasks) != 0 {
		test.Fatalf("Initial task fetch is not empty, found %d tasks.", len(tasks))
	}

	err = UpsertActiveTasks(testTasks)
	if err != nil {
		test.Fatalf("Failed to upsert initial tasks: '%v'.", err)
	}

	tasks, err = GetActiveTasks()
	if err != nil {
		test.Fatalf("Failed to fetch initial tasks: '%v'.", err)
	}

	if !reflect.DeepEqual(testTasks, tasks) {
		test.Fatalf("Initial tasks are not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(testTasks), util.MustToJSONIndent(tasks))
	}

	upsertTasks := tasks

	// Remove.
	upsertTasks["A"] = nil

	// Update Information.
	expectedNewCourseID := "ZZZ"
	upsertTasks["B"].CourseID = expectedNewCourseID

	// Ignore.
	delete(tasks, "C")
	delete(tasks, "D")

	err = UpsertActiveTasks(upsertTasks)
	if err != nil {
		test.Fatalf("Failed to upsert new tasks: '%v'.", err)
	}

	newTasks, err := GetActiveTasks()
	if err != nil {
		test.Fatalf("Failed to fetch new tasks: '%v'.", err)
	}

	expectedCount := len(testTasks) - 1
	actualCount := len(newTasks)
	if len(newTasks) != (len(testTasks) - 1) {
		test.Fatalf("New tasks hand unexpected count. Expected: %d, Actual: %d.", expectedCount, actualCount)
	}

	_, exists := newTasks["A"]
	if exists {
		test.Fatalf("Found task that should have been removed.")
	}

	if expectedNewCourseID != newTasks["B"].CourseID {
		test.Fatalf("Unexpected upserted course ID. Expected: '%s', Actual: '%s'.", expectedNewCourseID, newTasks["B"].CourseID)
	}
}

func (this *DBTests) DBTestGetActiveCourseTasksBase(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	course := MustGetTestCourse()

	err := UpsertActiveTasks(testTasks)
	if err != nil {
		test.Fatalf("Failed to upsert initial tasks: '%v'.", err)
	}

	tasks, err := GetActiveCourseTasks(course)
	if err != nil {
		test.Fatalf("Failed to fetch tasks: '%v'.", err)
	}

	expectedKeys := []string{"A", "C"}

	actualKeys := make([]string, 0)
	for key, _ := range tasks {
		actualKeys = append(actualKeys, key)
	}
	slices.Sort(actualKeys)

	if !reflect.DeepEqual(expectedKeys, actualKeys) {
		test.Fatalf("Did not get expected keys. Expected: '%v', Actual: '%v'.", expectedKeys, actualKeys)
	}
}

func (this *DBTests) DBTestDisableTask(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	testTask := &model.UserTaskInfo{
		Type:     model.TaskTypeTest,
		Name:     "test",
		Disabled: false,
		When:     &util.ScheduledTime{Daily: "00:00"},
		Options:  nil,
	}

	course := MustGetTestCourse()
	course.Tasks = []*model.UserTaskInfo{testTask}
	MustSaveCourse(course)

	tasks, err := GetActiveCourseTasks(course)
	if err != nil {
		test.Fatalf("Failed to fetch initial tasks: '%v'.", err)
	}

	if len(tasks) != 1 {
		test.Fatalf("Did not get expected number of initial tasks. Expected: %d, Actual: %d.", 1, len(tasks))
	}

	testTask.Disabled = true

	course = MustGetTestCourse()
	course.Tasks = []*model.UserTaskInfo{testTask}
	MustSaveCourse(course)

	tasks, err = GetActiveCourseTasks(course)
	if err != nil {
		test.Fatalf("Failed to fetch final tasks: '%v'.", err)
	}

	if len(tasks) != 0 {
		test.Fatalf("Did not get expected number of final tasks. Expected: %d, Actual: %d.", 0, len(tasks))
	}
}

func (this *DBTests) DBTestGetNextActiveTaskBase(test *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	err := UpsertActiveTasks(testTasks)
	if err != nil {
		test.Fatalf("Failed to upsert initial tasks: '%v'.", err)
	}

	expextedOrder := []string{"C", "A", "B", "D", ""}
	for i, expectedHash := range expextedOrder {
		task, err := GetNextActiveTask()
		if err != nil {
			test.Fatalf("Round %d: Failed to fetch next task: '%v'.", i, err)
		}

		if expectedHash == "" {
			if task != nil {
				test.Fatalf("Got a task on the last round, when it should have been nil: '%s'.", task.Hash)
			}

			continue
		}

		if expectedHash != task.Hash {
			test.Fatalf("Round %d: Did not get expected hash. Expected: '%s', Actual: '%s'.", i, expectedHash, task.Hash)
		}

		upsertTasks := map[string]*model.FullScheduledTask{expectedHash: nil}
		err = UpsertActiveTasks(upsertTasks)
		if err != nil {
			test.Fatalf("Failed to remove task: '%v'.", err)
		}
	}
}

// Note that the hashes and tasks are not real and only work if we don't validate them.
var testTasks map[string]*model.FullScheduledTask = map[string]*model.FullScheduledTask{
	"A": &model.FullScheduledTask{
		SystemTaskInfo: model.SystemTaskInfo{
			Source:      model.TaskSourceCourse,
			NextRunTime: timestamp.FromMSecs(100),
			CourseID:    TEST_COURSE_ID,
			Hash:        "A",
		},
	},
	"B": &model.FullScheduledTask{
		SystemTaskInfo: model.SystemTaskInfo{
			Source:      model.TaskSourceUnknown,
			NextRunTime: timestamp.FromMSecs(101),
			CourseID:    TEST_COURSE_ID,
			Hash:        "B",
		},
	},
	"C": &model.FullScheduledTask{
		SystemTaskInfo: model.SystemTaskInfo{
			Source:      model.TaskSourceCourse,
			NextRunTime: timestamp.FromMSecs(0),
			CourseID:    TEST_COURSE_ID,
			Hash:        "C",
		},
	},
	"D": &model.FullScheduledTask{
		SystemTaskInfo: model.SystemTaskInfo{
			Source:      model.TaskSourceUnknown,
			NextRunTime: timestamp.FromMSecs(103),
			Hash:        "D",
		},
	},
}
