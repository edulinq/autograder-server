package db

// The db package acts as an interface for other packages into the database.
// Once db.Open() is called, this layer will handle conversions to/from the databse.

import (
    "fmt"
    "sync"

    "github.com/rs/zerolog/log"

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
    EnsureTables() error;
    GetCourseUsers(courseID string) (map[string]*usr.User, error);

    // Get all known courses.
    GetCourses() (map[string]*types.Course, error);

    // Get a specific course.
    // Returns (nil, nil) if the course does not exist.
    GetCourse(courseID string) (*types.Course, error);

    // Load a course into the databse and return the course's id.
    // This implies loading a course directory from a config and saving it in the db.
    // Override any existing settings for this course.
    LoadCourse(path string) (string, error);

    // Explicitly save a course.
    SaveCourse(course *types.Course) error;

    // Explicitly save an assignment.
    SaveAssignment(assignment *types.Assignment) error;
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
