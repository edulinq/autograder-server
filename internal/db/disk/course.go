package disk

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

const DISK_DB_COURSES_DIR = "courses"

func (this *backend) ClearCourse(course *model.Course) error {
	path := this.getCoursePath(course)

	this.contextLock(path)
	defer this.contextUnlock(path)

	err := util.RemoveDirent(this.getCourseDir(course))
	if err != nil {
		return fmt.Errorf("Failed to remove course dir for '%s': '%w'.", course.GetID(), err)
	}

	err = this.clearCourseUsers(course)
	if err != nil {
		return fmt.Errorf("Failed to drop users from removed course: '%w'.", err)
	}

	return nil
}

func (this *backend) AddTestCourse(path string) (*model.Course, error) {
	path = util.ShouldAbs(path)

	this.contextLock(path)
	defer this.contextUnlock(path)

	course, submissions, err := model.FullLoadCourseFromPath(path, true)
	if err != nil {
		return nil, err
	}

	err = this.saveCourseLock(course, false)
	if err != nil {
		return nil, err
	}

	err = this.saveSubmissionsLock(course, submissions, false)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (this *backend) SaveCourse(course *model.Course) error {
	return this.saveCourseLock(course, true)
}

func (this *backend) saveCourseLock(course *model.Course, acquireLock bool) error {
	path := this.getCoursePath(course)

	if acquireLock {
		this.contextLock(path)
		defer this.contextUnlock(path)
	}

	util.MkDir(this.getCourseDir(course))

	err := util.ToJSONFileIndent(course, path)
	if err != nil {
		return err
	}

	for _, assignment := range course.Assignments {
		err = this.saveAssignmentLock(assignment, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *backend) DumpCourse(course *model.Course, targetDir string) error {
	path := this.getCoursePath(course)

	this.contextReadLock(path)
	defer this.contextReadUnlock(path)

	// Just directly copy the course's dir in the DB.
	err := util.CopyDirContents(this.getCourseDir(course), targetDir)
	if err != nil {
		return fmt.Errorf("Failed to copy disk db '%s' into '%s': '%w'.", this.baseDir, targetDir, err)
	}

	return nil
}

func (this *backend) GetCourse(courseID string) (*model.Course, error) {
	path := this.getCoursePathFromID(courseID)

	this.contextReadLock(path)
	defer this.contextReadUnlock(path)

	if !util.PathExists(path) {
		return nil, nil
	}

	return model.LoadCourseFromPath(path, false)
}

func (this *backend) GetCourses() (map[string]*model.Course, error) {
	coursesDir := filepath.Join(this.baseDir, DISK_DB_COURSES_DIR)

	configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, coursesDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to search for courses in '%s': '%w'.", coursesDir, err)
	}

	courses := make(map[string]*model.Course, len(configPaths))
	for _, configPath := range configPaths {
		err := func() error {
			configPath = util.ShouldAbs(configPath)

			this.contextReadLock(configPath)
			defer this.contextReadUnlock(configPath)

			course, err := model.LoadCourseFromPath(configPath, false)
			if err != nil {
				return fmt.Errorf("Failed to load course '%s': '%w'", configPath, err)
			}

			courses[course.GetID()] = course

			return nil
		}()

		if err != nil {
			return nil, err
		}
	}

	return courses, nil
}

func (this *backend) getCourseDir(course *model.Course) string {
	return this.getCourseDirFromID(course.GetID())
}

func (this *backend) getCoursePath(course *model.Course) string {
	return this.getCoursePathFromID(course.GetID())
}

func (this *backend) getCourseDirFromID(courseID string) string {
	path := filepath.Join(this.baseDir, DISK_DB_COURSES_DIR, courseID)
	return util.ShouldAbs(path)
}

func (this *backend) getCoursePathFromID(courseID string) string {
	return filepath.Join(this.getCourseDirFromID(courseID), model.COURSE_CONFIG_FILENAME)
}
