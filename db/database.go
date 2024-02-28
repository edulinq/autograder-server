package db

// The db package acts as an interface for other packages into the database.
// Once db.Open() is called, this layer will handle conversions to/from the databse.
// Any Get*() functions that return an interface will return a pure nil if nothing is found.
// When working with courses, Load*() functions are for courses that are already added to the system,
// use Add*() functions for new courses.

import (
    "fmt"
    "sync"
    "time"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db/disk"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/model"
)

var backend Backend;
var dbLock sync.Mutex;

const (
    DB_TYPE_DISK = "disk"
    DB_TYPE_SQLITE = "sqlite"
    DB_TYPE_POSTGRES = "postgres"
)

// Actual database implementations.
// Any ID (course, assignment, etc) passed into a backend will be sanitized.
type Backend interface {
    Close() error;
    Clear() error;
    EnsureTables() error;

    // Clear all information about a course.
    ClearCourse(course *model.Course) error;

    // Get all known courses.
    GetCourses() (map[string]*model.Course, error);

    // Get a specific course.
    // Returns (nil, nil) if the course does not exist.
    GetCourse(courseID string) (*model.Course, error);

    // Load a course into the database and return the course.
    // This implies loading a course directory from a config and saving it in the db.
    // Will search for and load any assignments, users, and submissions
    // located in the same directory tree.
    // Override any existing settings for this course.
    LoadCourse(path string) (*model.Course, error);

    // Explicitly save a course (which includes all course assignments).
    SaveCourse(course *model.Course) error;

    // Dump a database course to the standard directory layout
    // in the specified directory.
    // The target directory should not exist, or be empty.
    DumpCourse(course *model.Course, targetDir string) error;

    // Explicitly save an assignment.
    SaveAssignment(assignment *model.Assignment) error;

    GetUsers(course *model.Course) (map[string]*model.User, error);

    // Get a specific user.
    // Returns nil if no matching user exists.
    GetUser(course *model.Course, email string) (*model.User, error);

    // Upsert the given users.
    SaveUsers(course *model.Course, users map[string]*model.User) error;

    // Remove a user.
    // Do nothing and return nil if the user does not exist.
    RemoveUser(course *model.Course, email string) error;

    // Remove a submission.
    // Return a bool indicating whether the submission exists or not and an error if there is one.
    RemoveSubmission(assignment *model.Assignment, email string, submissionID string) (bool, error);

    // Save the results of grading.
    // All the submissions should be from this course.
    SaveSubmissions(course *model.Course, results []*model.GradingResult) error;

    // Get the next short submission ID.
    GetNextSubmissionID(assignment *model.Assignment, email string) (string, error);

    // Get a history of all submissions for this assignment and user.
    GetSubmissionHistory(assignment *model.Assignment, email string) ([]*model.SubmissionHistoryItem, error);

    // Get the results from a specific (or most recent) submission.
    // The submission ID will either be a short submission ID, or empty (if the most recent submission is to be returned).
    // Can return nil if the submission does not exist.
    GetSubmissionResult(assignment *model.Assignment, email string, shortSubmissionID string) (*model.GradingInfo, error);

    // Get all attempts for a specific user.
    GetSubmissionAttempts(assignment *model.Assignment, email string) ([]*model.GradingResult, error);

    // Get the scoring infos for an assignment for all users that match the given role.
    // A role of model.RoleUnknown means all users.
    // Users without a submission (but with a matching role) will be represented with a nil map value.
    // A nil map should only be returned on error.
    GetScoringInfos(assignment *model.Assignment, filterRole model.UserRole) (map[string]*model.ScoringInfo, error);

    // Get recent submission result for each user of the given role.
    // A role of model.RoleUnknown means all users.
    // Users without a submission (but with a matching role) will be represented with a nil map value.
    // A nil map should only be returned on error.
    GetRecentSubmissions(assignment *model.Assignment, filterRole model.UserRole) (map[string]*model.GradingInfo, error);

    // Get an overview of the recent submission result for each user of the given role.
    // A role of model.RoleUnknown means all users.
    // Users without a submission (but with a matching role) will be represented with a nil map value.
    // A nil map should only be returned on error.
    GetRecentSubmissionSurvey(assignment *model.Assignment, filterRole model.UserRole) (map[string]*model.SubmissionHistoryItem, error);

    // Get the results of a submission including files and grading output.
    GetSubmissionContents(assignment *model.Assignment, email string, shortSubmissionID string) (*model.GradingResult, error);

    // Get the contents of recent submission result for each user of the given role.
    // A role of model.RoleUnknown means all users.
    // Users without a submission (but with a matching role) will be represented with a nil map value.
    // A nil map should only be returned on error.
    GetRecentSubmissionContents(assignment *model.Assignment, filterRole model.UserRole) (map[string]*model.GradingResult, error);

    // Record that a task has been completed.
    // The DB is only required to keep the most recently completed task with the given course/ID.
    LogTaskCompletion(courseID string, taskID string, instance time.Time) error;

    // Get the last time a task with the given course/ID was completed.
    // Will return a zero time (time.Time{}).
    GetLastTaskCompletion(courseID string, taskID string) (time.Time, error);

    // DB backends will also be used as logging storage backends.
    log.StorageBackend

    // Get any logs that that match the specific requirements.
    // Each parameter (except for the log level) can be passed with a zero value, in which case it will not be used for filtering.
    GetLogRecords(level log.LogLevel, after time.Time, courseID string, assignmentID string, userID string) ([]*log.Record, error);
}

func Open() error {
    dbLock.Lock();
    defer dbLock.Unlock();

    if (backend != nil) {
        return nil;
    }

    var err error;
    dbType := config.DB_TYPE.Get();

    switch dbType {
        case DB_TYPE_DISK:
            backend, err = disk.Open();
        default:
            err = fmt.Errorf("Unknown database type: '%s'.", dbType);
    }

    if (err != nil) {
        return fmt.Errorf("Failed to open database: %w.", err);
    }

    log.SetStorageBackend(backend);

    return backend.EnsureTables();
}

func Close() error {
    dbLock.Lock();
    defer dbLock.Unlock();

    if (backend == nil) {
        return nil;
    }

    err := backend.Close();
    backend = nil;

    return err;
}

func Clear() error {
    if (backend == nil) {
        return nil;
    }

    return backend.Clear();
}

func MustOpen() {
    err := Open();
    if (err != nil) {
        log.Fatal("Failed to open db.", err);
    }
}

func MustClose() {
    err := Close();
    if (err != nil) {
        log.Fatal("Failed to close db.", err);
    }
}

func MustClear() {
    err := Clear();
    if (err != nil) {
        log.Fatal("Failed to clear db.", err);
    }
}
