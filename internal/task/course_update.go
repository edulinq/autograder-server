package task

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/model/tasks"
	"github.com/edulinq/autograder/internal/procedures/courses"
)

func RunCourseUpdateTask(course *model.Course, rawTask tasks.ScheduledTask) (bool, error) {
	task, ok := rawTask.(*tasks.CourseUpdateTask)
	if !ok {
		return false, fmt.Errorf("Task is not a CourseUpdateTask: %t (%v).", rawTask, rawTask)
	}

	if task.Disable {
		return true, nil
	}

	err := updateCourse(course)

	// Do not reschedule, all course tasks were already scheduled.
	return false, err
}

// See procetures.UpdateCourse().
func updateCourse(course *model.Course) error {
	options := courses.CourseUpsertOptions{
		ContextUser: db.MustGetRoot(),
	}

	_, err := courses.UpdateFromLocalSource(course, options)
	if err != nil {
		log.Error("Failed to update course.", err, course)
		return err
	}

	return nil
}
