package disk

import (
    "fmt"
    "path/filepath"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/util"
)

const DISK_DB_COURSES_DIR = "courses";

func (this *backend) ClearCourse(course *model.Course) error {
    courseDir := this.getCourseDir(course);
    courseDir = util.ShouldAbs(courseDir);

    common.Lock(courseDir);
    defer common.Unlock(courseDir);

    err := util.RemoveDirent(courseDir);
    if (err != nil) {
        return fmt.Errorf("Failed to remove course dir for '%s': '%w'.", course.GetID(), err);
    }

    return nil;
}

func (this *backend) LoadCourse(path string) (*model.Course, error) {
    courseDir := util.ShouldAbs(path);

    common.Lock(courseDir);
    defer common.Unlock(courseDir);

    course, users, submissions, err := model.FullLoadCourseFromPath(courseDir);
    if (err != nil) {
        return nil, err;
    }

    log.Debug("Loaded disk course.",
            log.NewAttr("database", "disk"), log.NewAttr("path", courseDir),
            log.NewAttr("id", course.GetID()), log.NewAttr("num-assignments", len(course.Assignments)));

    err = this.saveCourseLock(course, false);
    if (err != nil) {
        return nil, err;
    }

    err = this.saveUsersLock(course, users, false);
    if (err != nil) {
        return nil, err;
    }

    err = this.saveSubmissionsLock(course, submissions, false);
    if (err != nil) {
        return nil, err;
    }

    return course, nil;
}

func (this *backend) SaveCourse(course *model.Course) error {
    return this.saveCourseLock(course, true);
}

func (this *backend) saveCourseLock(course *model.Course, acquireLock bool) error {
    courseDir := this.getCourseDir(course);
    courseDir = util.ShouldAbs(courseDir);

    if (acquireLock) {
        common.Lock(courseDir);
        defer common.Unlock(courseDir);
    }

    util.MkDir(courseDir);

    err := util.ToJSONFileIndent(course, this.getCoursePath(course));
    if (err != nil) {
        return err;
    }

    for _, assignment := range course.Assignments {
        err = this.saveAssignmentLock(assignment, false);
        if (err != nil) {
            return err;
        }
    }

    return nil;
}

func (this *backend) DumpCourse(course *model.Course, targetDir string) error {
    courseDir := this.getCourseDir(course);
    courseDir = util.ShouldAbs(courseDir);

    common.ReadLock(courseDir);
    defer common.ReadUnlock(courseDir);

    // Just directly copy the course's dir in the DB.
    err := util.CopyDirContents(courseDir, targetDir);
    if (err != nil) {
        return fmt.Errorf("Failed to copy disk db '%s' into '%s': '%w'.", this.baseDir, targetDir, err);
    }

    return nil;
}

func (this *backend) GetCourse(courseID string) (*model.Course, error) {
    courseDir := this.getCourseDirFromID(courseID);
    courseDir = util.ShouldAbs(courseDir);

    common.ReadLock(courseDir);
    defer common.ReadUnlock(courseDir);

    path := this.getCoursePathFromID(courseID);
    if (!util.PathExists(path)) {
        return nil, nil;
    }

    return model.LoadCourseFromPath(path);
}

func (this *backend) GetCourses() (map[string]*model.Course, error) {
    coursesDir := filepath.Join(this.baseDir, DISK_DB_COURSES_DIR);

    configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, coursesDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for courses in '%s': '%w'.", coursesDir, err);
    }

    courses := make(map[string]*model.Course, len(configPaths));
    for _, configPath := range configPaths {
        err := func() error {
            configPath = util.ShouldAbs(configPath);

            common.ReadLock(configPath);
            defer common.ReadUnlock(configPath);

            course, err := model.LoadCourseFromPath(configPath);
            if (err != nil) {
                return fmt.Errorf("Failed to load course '%s': '%w'", configPath, err);
            }
            
            courses[course.GetID()] = course;
            
            return nil;
        }();

        if (err != nil) {
            return nil, err;
        }
    }

    return courses, nil;
}

func (this *backend) getCourseDir(course *model.Course) string {
    return this.getCourseDirFromID(course.GetID());
}

func (this *backend) getCoursePath(course *model.Course) string {
    return this.getCoursePathFromID(course.GetID());
}

func (this *backend) getCourseDirFromID(courseID string) string {
    return filepath.Join(this.baseDir, DISK_DB_COURSES_DIR, courseID);
}

func (this *backend) getCoursePathFromID(courseID string) string {
    return filepath.Join(this.getCourseDirFromID(courseID), model.COURSE_CONFIG_FILENAME);
}
