package tasks

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/scoring"
)

func RunCourseScoringUploadTask(task *model.FullScheduledTask) error {
	course, err := db.GetCourse(task.CourseID)
	if err != nil {
		return fmt.Errorf("Failed to get course '%s': '%w'.", task.CourseID, err)
	}

	if course == nil {
		return fmt.Errorf("Unable to find course '%s'.", task.CourseID)
	}

	_, err = scoring.FullCourseScoringAndUpload(course, false)
	return err
}
