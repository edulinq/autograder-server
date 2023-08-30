package grader

import (
    "fmt"
    "path/filepath"
    "sync"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

const GRADING_INPUT_DIRNAME = "input"
const GRADING_OUTPUT_DIRNAME = "output"
const GRADING_WORK_DIRNAME = "work"

// TODO(eriq): Create a maintenance task that removes old, unused locks.
var submissionLocks sync.Map;

type GradeOptions struct {
    NoDocker bool
    LeaveTempDir bool
}

func GetDefaultGradeOptions() GradeOptions {
    return GradeOptions{
        NoDocker: config.GetBool(config.DOCKER_DISABLE),
    };
}

// Grade with default options pulled from config.
func GradeDefault(assignment *model.Assignment, submissionPath string, user string) (*model.GradedAssignment, error) {
    return Grade(assignment, submissionPath, user, GetDefaultGradeOptions());
}

// Grade with custom options.
func Grade(assignment *model.Assignment, submissionPath string, user string, options GradeOptions) (*model.GradedAssignment, error) {
    lockKey := fmt.Sprintf("%s::%s::%s", assignment.Course.ID, assignment.ID, user);

    // Get the existing mutex, or store (and fetch) a new one.
    val, _ := submissionLocks.LoadOrStore(lockKey, &sync.Mutex{});
    lock := val.(*sync.Mutex)

    lock.Lock();
    defer lock.Unlock();

    submissionDir, err := assignment.Course.PrepareSubmission(user);
    if (err != nil) {
        return nil, err;
    }

    // TODO(eriq): Copy the submission to the user's submission directory.

    outputDir := filepath.Join(submissionDir, GRADING_OUTPUT_DIRNAME);

    if (options.NoDocker) {
        return RunNoDockerGrader(assignment, submissionPath, outputDir, options);
    }

    return RunDockerGrader(assignment, submissionPath, outputDir, options);
}
