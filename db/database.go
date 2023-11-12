package db

// The db package acts as an interface for other packages into the database.
// Once db.Open() is called, this layer will handle conversions to/from the databse.
// Any Get*() functions that return an interface will return a pure nil if nothing is found.

import (
    "fmt"
    "sync"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db/disk"
    "github.com/eriq-augustine/autograder/db/types"
    "github.com/eriq-augustine/autograder/usr"
)

var backend Backend;
var dbLock sync.Mutex;

const (
    DB_TYPE_DISK = "disk"
    DB_TYPE_SQLITE = "sqlite"
    DB_TYPE_POSTGRES = "postgres"
)

// Actual databse implementations.
// Backends will always deal with concrete types from db/types and not interfaces from model.
// Any ID (course, assignment, etc) passed into a backend will be sanitized.
type Backend interface {
    Close() error;
    Clear() error;
    EnsureTables() error;

    // Get all known courses.
    GetCourses() (map[string]*types.Course, error);

    // Get a specific course.
    // Returns (nil, nil) if the course does not exist.
    GetCourse(courseID string) (*types.Course, error);

    // Load a course into the databse and return the course's id.
    // This implies loading a course directory from a config and saving it in the db.
    // Will search for and load any assignments, users, and submissions
    // located in the same directory tree.
    // Override any existing settings for this course.
    LoadCourse(path string) (string, error);

    // Explicitly save a course.
    SaveCourse(course *types.Course) error;

    // Explicitly save an assignment.
    SaveAssignment(assignment *types.Assignment) error;

    GetUsers(course *types.Course) (map[string]*usr.User, error);

    // Get a specific user.
    // Returns nil if no matching user exists.
    GetUser(course *types.Course, email string) (*usr.User, error);

    // Upsert the given users.
    SaveUsers(course *types.Course, users map[string]*usr.User) error;

    // Remove a user.
    // Do nothing and return nil if the user does not exist.
    RemoveUser(course *types.Course, email string) error;

    // Save the results of grading.
    // All the submissions should be from this course.
    SaveSubmissions(course *types.Course, results []*artifact.GradingResult) error;

    // Get the next short submission ID.
    GetNextSubmissionID(assignment *types.Assignment, email string) (string, error);

    // Get a history of all submissions for this assignment and user.
    GetSubmissionHistory(assignment *types.Assignment, email string) ([]*artifact.SubmissionHistoryItem, error);

    // Get the results from a specific (or most recent) submission.
    // The submission ID will either be a short submission ID, or empty (if the most recent submission is to be returned).
    // Can return nil if the submission does not exist.
    GetSubmissionResult(assignment *types.Assignment, email string, shortSubmissionID string) (*artifact.GradingInfo, error);

    // Get the scoring infos for an assignment for all users that match the given role.
    // A role of usr.Unknown means all users.
    // Users without a submission (but with a matching role) will be represented with a nil map value.
    GetScoringInfos(assignment *types.Assignment, filterRole usr.UserRole) (map[string]*artifact.ScoringInfo, error);

    // Get recent submission result for each user of the given role.
    // A role of usr.Unknown means all users.
    // Users without a submission (but with a matching role) will be represented with a nil map value.
    GetRecentSubmissions(assignment *types.Assignment, filterRole usr.UserRole) (map[string]*artifact.GradingInfo, error);

    // Get an overview of the recent submission result for each user of the given role.
    // A role of usr.Unknown means all users.
    // Users without a submission (but with a matching role) will be represented with a nil map value.
    GetRecentSubmissionSurvey(assignment *types.Assignment, filterRole usr.UserRole) (map[string]*artifact.SubmissionHistoryItem, error);

    // Get the results of a submission including files and grading output.
    GetSubmissionContents(assignment *types.Assignment, email string, shortSubmissionID string) (*artifact.GradingResult, error);

    // Get the contents of recent submission result for each user of the given role.
    // A role of usr.Unknown means all users.
    // Users without a submission (but with a matching role) will be represented with a nil map value.
    GetRecentSubmissionContents(assignment *types.Assignment, filterRole usr.UserRole) (map[string]*artifact.GradingResult, error);
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
        log.Fatal().Err(err).Msg("Failed to open db.");
    }
}

func MustClose() {
    err := Close();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to close db.");
    }
}

func MustClear() {
    err := Clear();
    if (err != nil) {
        log.Fatal().Err(err).Msg("Failed to clear db.");
    }
}
