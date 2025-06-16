package db

// The db package acts as an interface for other packages into the database.
// Once db.Open() is called, this layer will handle conversions to/from the databse.
// Any Get*() functions that return an interface will return a pure nil if nothing is found.
// When working with courses, Load*() functions are for courses that are already added to the system,
// use Add*() functions for new courses.

import (
	"fmt"
	"sync"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db/disk"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/stats"
)

var backend Backend
var dbLock sync.Mutex
var loadTestData bool

const (
	DB_TYPE_DISK     = "disk"
	DB_TYPE_SQLITE   = "sqlite"
	DB_TYPE_POSTGRES = "postgres"
)

// Actual database implementations.
// Any ID (course, assignment, etc) passed into a backend will be sanitized.
// Note that the database package provides more functionality than what is provided directly by a Backend.
type Backend interface {
	// Administrative Operations

	Close() error
	Clear() error
	EnsureTables() error

	// Course Operations

	// Clear all information about a course.
	ClearCourse(course *model.Course) error

	// Get all known courses.
	GetCourses() (map[string]*model.Course, error)

	// Get a specific course.
	// Returns (nil, nil) if the course does not exist.
	GetCourse(courseID string) (*model.Course, error)

	// Add a course to the database specifically for testing purposes.
	// This course may include test submissions that must also be added.
	AddTestCourse(path string) (*model.Course, error)

	// Explicitly save a course (which includes all course assignments).
	SaveCourse(course *model.Course) error

	// Dump a database course to the standard directory layout
	// in the specified directory.
	// The target directory should not exist, or be empty.
	DumpCourse(course *model.Course, targetDir string) error

	// Assignment Operations

	// Explicitly save an assignment.
	SaveAssignment(assignment *model.Assignment) error

	// User Operations
	// User maps always map the user's ID to an actual user pointer.

	// Get all the users on the server.
	GetServerUsers() (map[string]*model.ServerUser, error)

	// Get all the users in a course.
	GetCourseUsers(course *model.Course) (map[string]*model.CourseUser, error)

	// Get a specific server user.
	// Returns nil if no matching user exists.
	GetServerUser(email string) (*model.ServerUser, error)

	// Upsert the given users.
	// Only fields to be updated should be non-nil (excluding email, which will always exist).
	// Any group membership (roles) will be added to. A missing role does not imply deletion.
	UpsertUsers(users map[string]*model.ServerUser) error

	// Remove a user from the server.
	// Do nothing and return nil if the user does not exist.
	DeleteUser(email string) error

	// Remove a user from a course (but leave the user on the server).
	// Do nothing and return nil if the user does not exist in that course.
	RemoveUserFromCourse(course *model.Course, email string) error

	// Delete a token owned by a user.
	// Return true if the user and token exists and the token was removed, false otherwise.
	DeleteUserToken(email string, tokenID string) (bool, error)

	// Submission Operations

	// Remove a submission.
	// Return a bool indicating whether the submission exists or not and an error if there is one.
	RemoveSubmission(assignment *model.Assignment, email string, submissionID string) (bool, error)

	// Save the results of grading.
	// All the submissions should be from this course.
	SaveSubmissions(course *model.Course, results []*model.GradingResult) error

	// Get the next short submission ID.
	// This will be the most recent ID for the user on this assignment.
	GetNextSubmissionID(assignment *model.Assignment, email string) (string, error)

	// Get the submission ID of the submission before this one (or empty string if this is the first one).
	GetPreviousSubmissionID(assignment *model.Assignment, email string, shortSubmissionID string) (string, error)

	// Get a history of all submissions for this assignment and user.
	GetSubmissionHistory(assignment *model.Assignment, email string) ([]*model.SubmissionHistoryItem, error)

	// Get the results from a specific (or most recent) submission.
	// The submission ID will either be a short submission ID, or empty (if the most recent submission is to be returned).
	// Can return nil if the submission does not exist.
	GetSubmissionResult(assignment *model.Assignment, email string, shortSubmissionID string) (*model.GradingInfo, error)

	// Get all attempts for a specific user.
	GetSubmissionAttempts(assignment *model.Assignment, email string) ([]*model.GradingResult, error)

	// TODO: (Lucas) Update function descriptions.
	// Get the scoring infos for an assignment for all users that match the given role.
	// A role of model.CourseRoleUnknown means all users.
	// Users without a submission (but with a matching role) will be represented with a nil map value.
	// A nil map should only be returned on error.
	GetScoringInfos(assignment *model.Assignment, reference model.ParsedCourseUserReference) (map[string]*model.ScoringInfo, error)

	// TODO: (Lucas) Update function descriptions.
	// Get recent submission result for each user of the given role.
	// A role of model.CourseRoleUnknown means all users.
	// Users without a submission (but with a matching role) will be represented with a nil map value.
	// A nil map should only be returned on error.
	GetRecentSubmissions(assignment *model.Assignment, reference model.ParsedCourseUserReference) (map[string]*model.GradingInfo, error)

	// TODO: (Lucas) Update function descriptions.
	// Get an overview of the recent submission result for each user of the given role.
	// A role of model.CourseRoleUnknown means all users.
	// Users without a submission (but with a matching role) will be represented with a nil map value.
	// A nil map should only be returned on error.
	GetRecentSubmissionSurvey(assignment *model.Assignment, reference model.ParsedCourseUserReference) (map[string]*model.SubmissionHistoryItem, error)

	// Get the results of a submission including files and grading output.
	GetSubmissionContents(assignment *model.Assignment, email string, shortSubmissionID string) (*model.GradingResult, error)

	// TODO: (Lucas) Update function descriptions.
	// Get the contents of recent submission result for each user of the given role.
	// A role of model.CourseRoleUnknown means all users.
	// Users without a submission (but with a matching role) will be represented with a nil map value.
	// A nil map should only be returned on error.
	GetRecentSubmissionContents(assignment *model.Assignment, reference model.ParsedCourseUserReference) (map[string]*model.GradingResult, error)

	// Task Operations

	// Get all the active tasks that come from the given course.
	// The returned tasks will be keyed by the task's hash.
	GetActiveCourseTasks(course *model.Course) (map[string]*model.FullScheduledTask, error)

	// Get all the active tasks keyed by hash.
	GetActiveTasks() (map[string]*model.FullScheduledTask, error)

	// Get the next active task that should be run.
	// Return nil if there are no active tasks.
	GetNextActiveTask() (*model.FullScheduledTask, error)

	// Upsert the given tasks.
	// The map of tasks is keyed by the task's hash (like with GetActiveCourseTasks()),
	// and a nil value indicates that the given task should be removed.
	UpsertActiveTasks(tasks map[string]*model.FullScheduledTask) error

	// Logging Operations

	// DB backends will also be used as logging storage backends.
	log.StorageBackend

	// Get any logs that match the specific requirements.
	// Each parameter (except for the log level) can be passed with a zero value, in which case it will not be used for filtering.
	GetLogRecords(query log.ParsedLogQuery) ([]*log.Record, error)

	// Stats Operations

	// DB backends will also be used as stats storage backends.
	stats.StorageBackend

	// Analysis Operations

	// Fetch any matching individual analysis results.
	// Any id not matched in the DB will not be represented in the output.
	GetIndividualAnalysis(fullSubmissionIDs []string) (map[string]*model.IndividualAnalysis, error)

	// Fetch any matching pairwise analysis results.
	// Any key not matched in the DB will not be represented in the output.
	GetPairwiseAnalysis(keys []model.PairwiseKey) (map[model.PairwiseKey]*model.PairwiseAnalysis, error)

	// Remove any matching individual analysis results.
	// The database is allowed to sort the passed in slice.
	RemoveIndividualAnalysis(fullSubmissionIDs []string) error

	// Remove any matching pairwise analysis results.
	// The database is allowed to sort the passed in slice.
	RemovePairwiseAnalysis(keys []model.PairwiseKey) error

	// Store the results of a individual analysis.
	StoreIndividualAnalysis(records []*model.IndividualAnalysis) error

	// Store the results of a pairwise analysis.
	StorePairwiseAnalysis(records []*model.PairwiseAnalysis) error
}

func Open() error {
	dbLock.Lock()
	defer dbLock.Unlock()

	if backend != nil {
		return nil
	}

	var err error
	dbType := config.DB_TYPE.Get()

	switch dbType {
	case DB_TYPE_DISK:
		backend, err = disk.Open()
	default:
		err = fmt.Errorf("Unknown database type: '%s'.", dbType)
	}

	if err != nil {
		return fmt.Errorf("Failed to open database: '%w'.", err)
	}

	// Initialize the logging backend as soon as possible.
	log.SetStorageBackend(backend)

	err = backend.EnsureTables()
	if err != nil {
		return err
	}

	err = initializeRootUser()
	if err != nil {
		return fmt.Errorf("Failed to initialize the root user: '%w'.", err)
	}

	// We are probably running unit tests, load the test data.
	if config.LOAD_TEST_DATA.Get() {
		err = addTestCourses()
		if err != nil {
			return fmt.Errorf("Failed to load test courses: '%w'.", err)
		}

		err = addTestUsers()
		if err != nil {
			return fmt.Errorf("Failed to load test users: '%w'.", err)
		}
	}

	// Initialize the stat storage backend when we are sure this backend is ready.
	stats.SetStorageBackend(backend)

	return nil
}

func Close() error {
	dbLock.Lock()
	defer dbLock.Unlock()

	if backend == nil {
		return nil
	}

	err := backend.Close()
	backend = nil

	// Remove the other uses of the database backend.
	log.SetStorageBackend(nil)
	stats.SetStorageBackend(nil)

	return err
}

func Clear() error {
	if backend == nil {
		return nil
	}

	err := backend.Clear()
	if err != nil {
		return err
	}

	err = initializeRootUser()
	if err != nil {
		return err
	}

	return nil
}

func MustOpen() {
	err := Open()
	if err != nil {
		log.Fatal("Failed to open db.", err)
	}
}

func MustClose() {
	err := Close()
	if err != nil {
		log.Fatal("Failed to close db.", err)
	}
}

func MustClear() {
	err := Clear()
	if err != nil {
		log.Fatal("Failed to clear db.", err)
	}
}
