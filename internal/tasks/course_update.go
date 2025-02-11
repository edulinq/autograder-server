package tasks

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/courses"
)

func RunCourseUpdateTask(task *model.FullScheduledTask) error {
	course, err := db.GetCourse(task.CourseID)
	if err != nil {
		return fmt.Errorf("Failed to get course '%s': '%w'.", task.CourseID, err)
	}

	if course == nil {
		return fmt.Errorf("Unable to find course '%s'.", task.CourseID)
	}

	options := courses.CourseUpsertOptions{
		ContextUser: db.MustGetRoot(),
	}

	_, err = courses.UpdateFromLocalSource(course, options)
	return err
}
