package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func ClearCourse(course *model.Course) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.ClearCourse(course)
}

// Get a course.
func GetCourse(rawCourseID string) (*model.Course, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	courseID, err := common.ValidateID(rawCourseID)
	if err != nil {
		return nil, fmt.Errorf("Failed to validate course id '%s': '%w'.", rawCourseID, err)
	}

	course, err := backend.GetCourse(courseID)
	if err != nil {
		return nil, err
	}

	if course == nil {
		return nil, nil
	}

	return course, nil
}

// Get a course or panic.
// This is a convenience function for the CLI mains that need to get a course.
func MustGetCourse(rawCourseID string) *model.Course {
	course, err := GetCourse(rawCourseID)
	if err != nil {
		log.Fatal("Failed to get course.", err, log.NewCourseAttr(rawCourseID))
	}

	if course == nil {
		log.Fatal("Could not find course.", log.NewCourseAttr(rawCourseID))
	}

	return course
}

// Get all the known courses.
func GetCourses() (map[string]*model.Course, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	dbCourses, err := backend.GetCourses()
	if err != nil {
		return nil, err
	}

	courses := make(map[string]*model.Course, len(dbCourses))
	for key, value := range dbCourses {
		courses[key] = value
	}

	return courses, nil
}

// Get all the known courses or panic.
// This is a convenience function for the CLI mains.
func MustGetCourses() map[string]*model.Course {
	courses, err := GetCourses()
	if err != nil {
		log.Fatal("Failed to get courses.", err)
	}

	return courses
}

func SaveCourse(course *model.Course) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	err := course.Validate()
	if err != nil {
		return fmt.Errorf("Course '%s' is not valid: '%w'.", course.GetID(), err)
	}

	err = backend.SaveCourse(course)
	if err != nil {
		return err
	}

	// Upsert any active tasks.
	err = upsertActiveCourseTasks(course)
	if err != nil {
		return fmt.Errorf("Failed to upsert active tasks for course '%s': '%w'.", course.GetID(), err)
	}

	return nil
}

func upsertActiveCourseTasks(course *model.Course) error {
	// Get the current active tasks for this course.
	oldTasks, err := backend.GetActiveCourseTasks(course)
	if err != nil {
		return err
	}

	// Start with the assumption that we will remove all active tasks.
	newTasks := make(map[string]*model.FullScheduledTask, len(oldTasks))
	for hash, _ := range oldTasks {
		newTasks[hash] = nil
	}

	// Add in any new tasks and merge with any exiting tasks.
	for i, task := range course.Tasks {
		newTask, err := task.ToFullCourseTask(course.GetID())
		if err != nil {
			return fmt.Errorf("Unable to upsert task at inded %d: '%w'.", i, err)
		}

		if newTask == nil {
			continue
		}

		newTask.MergeTimes(oldTasks[newTask.Hash])
		newTasks[newTask.Hash] = newTask
	}

	// Save.
	if len(newTasks) > 0 {
		err = backend.UpsertActiveTasks(newTasks)
	}

	return err
}

func MustSaveCourse(course *model.Course) {
	err := SaveCourse(course)
	if err != nil {
		log.Fatal("Failed to save course.", err)
	}
}

func DumpCourse(course *model.Course, targetDir string) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	if !util.PathExists(targetDir) {
		err := util.MkDir(targetDir)
		if err != nil {
			return err
		}
	}

	if !util.IsEmptyDir(targetDir) {
		return fmt.Errorf("Dump target dir '%s' is not empty.", targetDir)
	}

	return backend.DumpCourse(course, targetDir)
}
