package disk

import (
    "fmt"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/util"
)

const DISK_DB_COURSES_DIR = "courses";

func (this *backend) LoadCourse(path string) error {
    this.lock.Lock();
    defer this.lock.Unlock();

    course, err := types.LoadCourse(path);
    if (err != nil) {
        return err;
    }

    log.Info().Str("database", "disk").Str("path", path).Str("id", course.GetID()).
            Int("num-assignments", len(course.Assignments)).Msg("Loaded course.");

    return this.saveCourseLock(course, false);
}

func (this *backend) SaveCourse(course *types.Course) error {
    return this.saveCourseLock(course, true);
}

func (this *backend) saveCourseLock(course *types.Course, acquireLock bool) error {
    if (acquireLock) {
        this.lock.Lock();
        defer this.lock.Unlock();
    }

    util.MkDir(this.getCourseDir(course));

    err := util.ToJSONFile(course, this.getCoursePath(course));
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

func (this *backend) GetCourse(courseID string) (*types.Course, error) {
    this.lock.Lock();
    defer this.lock.Unlock();

    path := this.getCoursePathFromID(courseID);
    if (!util.PathExists(path)) {
        return nil, nil;
    }

    return types.LoadCourse(path);
}

func (this *backend) GetCourses() (map[string]*types.Course, error) {
    this.lock.Lock();
    defer this.lock.Unlock();

    coursesDir := filepath.Join(this.baseDir, DISK_DB_COURSES_DIR);

    configPaths, err := util.FindFiles(types.COURSE_CONFIG_FILENAME, coursesDir);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to search for courses in '%s': '%w'.", coursesDir, err);
    }

    courses := make(map[string]*types.Course, len(configPaths));
    for _, configPath := range configPaths {
        course, err := types.LoadCourse(configPath);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to load course '%s': '%w'.", configPath, err);
        }

        courses[course.GetID()] = course;
    }

    return courses, nil;
}

func (this *backend) getCourseDir(course *types.Course) string {
    return this.getCourseDirFromID(course.GetID());
}

func (this *backend) getCoursePath(course *types.Course) string {
    return this.getCoursePathFromID(course.GetID());
}

func (this *backend) getCourseDirFromID(courseID string) string {
    return filepath.Join(this.baseDir, DISK_DB_COURSES_DIR, courseID);
}

func (this *backend) getCoursePathFromID(courseID string) string {
    return filepath.Join(this.getCourseDirFromID(courseID), DISK_DB_COURSES_DIR);
}
