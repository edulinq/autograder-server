package disk

import (
    "fmt"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const DISK_DB_COURSES_DIR = "courses";

func (this *backend) ClearCourse(course *model.Course) error {
    this.lock.Lock();
    defer this.lock.Unlock();

    err := util.RemoveDirent(this.getCourseDir(course));
    if (err != nil) {
        return fmt.Errorf("Failed to remove course dir for '%s': '%w'.", course.GetID(), err);
    }

    return nil;
}

func (this *backend) LoadCourse(path string) (*model.Course, error) {
    this.lock.Lock();
    defer this.lock.Unlock();

    course, users, submissions, err := model.FullLoadCourseFromPath(path);
    if (err != nil) {
        return nil, err;
    }

    log.Info().Str("database", "disk").Str("path", path).Str("id", course.GetID()).
            Int("num-assignments", len(course.Assignments)).Msg("Loaded course.");

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
    if (acquireLock) {
        this.lock.Lock();
        defer this.lock.Unlock();
    }

    util.MkDir(this.getCourseDir(course));

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
    this.lock.RLock();
    defer this.lock.RUnlock();

    // Just directly copy the course's dir in the DB.
    err := util.CopyDirContents(this.getCourseDir(course), targetDir);
    if (err != nil) {
        return fmt.Errorf("Failed to copy disk db '%s' into '%s': '%w'.", this.baseDir, targetDir, err);
    }

    return nil;
}

func (this *backend) GetCourse(courseID string) (*model.Course, error) {
    this.lock.RLock();
    defer this.lock.RUnlock();

    path := this.getCoursePathFromID(courseID);
    if (!util.PathExists(path)) {
        return nil, nil;
    }

    return model.LoadCourseFromPath(path);
}

func (this *backend) GetCourses() (map[string]*model.Course, error) {
    this.lock.RLock();
    defer this.lock.RUnlock();

    coursesDir := filepath.Join(this.baseDir, DISK_DB_COURSES_DIR);

    configPaths, err := util.FindFiles(model.COURSE_CONFIG_FILENAME, coursesDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for courses in '%s': '%w'.", coursesDir, err);
    }

    courses := make(map[string]*model.Course, len(configPaths));
    for _, configPath := range configPaths {
        course, err := model.LoadCourseFromPath(configPath);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to load course '%s': '%w'.", configPath, err);
        }

        courses[course.GetID()] = course;
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
