package db

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
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

// Load a course into the database from an existing path.
func LoadCourse(path string) (*model.Course, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	course, err := backend.LoadCourse(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to load course from path '%s': '%w'.", path, err)
	}

	err = course.Validate()
	if err != nil {
		return nil, fmt.Errorf("Failed to validate course from path '%s': '%w'.", path, err)
	}

	log.Debug("Loaded course.", course, log.NewAttr("path", path), log.NewAttr("num-assignments", len(course.Assignments)))

	return course, nil
}

func SaveCourse(course *model.Course) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.SaveCourse(course)
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

// Search the courses root directory and add all the associated courses and assignments.
// Return all the loaded course ids.
func AddCourses() ([]string, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return AddCoursesFromDir(config.GetCourseImportDir(), nil)
}

func MustAddCourses() []string {
	courseIDs, err := AddCourses()
	if err != nil {
		log.Fatal("Failed to load courses.", err, log.NewAttr("path", config.GetCourseImportDir()))
	}

	return courseIDs
}

func AddCoursesFromDir(baseDir string, source *common.FileSpec) ([]string, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err)
	}

	log.Debug("Number of importable course configs found.", log.NewAttr("count", len(configPaths)), log.NewAttr("dir", baseDir))

	courseIDs := make([]string, 0, len(configPaths))
	for _, configPath := range configPaths {
		course, err := AddCourse(configPath, source)
		if err != nil {
			return nil, fmt.Errorf("Could not load course '%s': '%w'.", configPath, err)
		}

		courseIDs = append(courseIDs, course.GetID())
	}

	return courseIDs, nil
}

// Add a course to the db from a path.
func AddCourse(path string, source *common.FileSpec) (*model.Course, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	course, err := LoadCourse(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to load course config '%s': '%w'.", path, err)
	}

	update := true
	saveCourse := false

	// Use the override source if not nil.
	if source != nil {
		course.Source = source
		saveCourse = true
	}

	// If the course's source is empty, set it to this directory where it is being added from.
	if (course.Source == nil) || course.Source.IsEmpty() {
		course.Source = common.GetPathFileSpec(util.ShouldAbs(filepath.Dir(path)))
		err = course.Source.Validate()
		if err != nil {
			return nil, fmt.Errorf("Failed to create source FileSpec: '%w'.", err)
		}

		saveCourse = true
		update = false
	}

	if saveCourse {
		err = SaveCourse(course)
		if err != nil {
			return nil, fmt.Errorf("Failed to save course: '%w'.", err)
		}
	}

	if !update {
		return course, nil
	}

	// Try to update the course from source.

	newCourse, _, err := UpdateCourseFromSource(course)
	if err != nil {
		return nil, err
	}

	return newCourse, nil
}

func MustAddCourse(path string) *model.Course {
	course, err := AddCourse(path, nil)
	if err != nil {
		log.Fatal("Failed to add course.", err, log.NewAttr("path", path))
	}

	return course
}

// Get a fresh copy of the course from the source and load it into the DB
// (thereby updating the course).
// The new course (or old course if no update happens) will be returned.
// The boolean return indicates if an update attempt was made.
// Callers to this should consider if tasks should be stopped before,
// and if tasks should be started and images rebuilt after.
func UpdateCourseFromSource(course *model.Course) (*model.Course, bool, error) {
	if backend == nil {
		return nil, false, fmt.Errorf("Database has not been opened.")
	}

	source := course.GetSource()

	if (source == nil) || source.IsEmpty() || source.IsNil() {
		return course, false, nil
	}

	baseDir := course.GetBaseSourceDir()

	if util.PathExists(baseDir) {
		err := util.RemoveDirent(baseDir)
		if err != nil {
			return nil, false, fmt.Errorf("Failed to remove existing course base source output '%s': '%w'.", baseDir, err)
		}
	}

	err := util.MkDir(baseDir)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to make course base source dir '%s': '%w'.", baseDir, err)
	}

	err = source.CopyTarget(common.ShouldGetCWD(), baseDir, false)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to copy course source ('%s') into course base source dir ('%s'): '%w'.", source, baseDir, err)
	}

	configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, baseDir)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to search for course configs in '%s': '%w'.", baseDir, err)
	}

	if len(configPaths) == 0 {
		return nil, false, fmt.Errorf("Did not find any course configs in course source ('%s'), should be exactly one.", source)
	}

	if len(configPaths) > 1 {
		return nil, false, fmt.Errorf("Found too many course configs (%d) in course source ('%s'), should be exactly one.", len(configPaths), source)
	}

	configPath := util.ShouldAbs(configPaths[0])

	newCourse, err := LoadCourse(configPath)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to load updated course: '%w'.", err)
	}

	// Ensure that the source is passed along.
	// This can happen when a course is loaded from a directory (without a source).
	if (newCourse.Source == nil) || newCourse.Source.IsEmpty() {
		newCourse.Source = source

		err = SaveCourse(newCourse)
		if err != nil {
			return nil, false, fmt.Errorf("Failed to save new course: '%w'.", err)
		}
	}

	return newCourse, true, nil
}
