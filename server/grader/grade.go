package grader

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const GRADING_INPUT_DIRNAME = "input"
const GRADING_OUTPUT_DIRNAME = "output"
const GRADING_WORK_DIRNAME = "work"

// TODO(eriq): Create a maintenance task that removes old, unused locks.
var submissionLocks sync.Map;

type GradeOptions struct {
    UseFakeSubmissionsDir bool
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

    submissionDir, err := prepSubmissionDir(assignment, user, options);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to prepare submission dir for assignment '%s' and user '%s': '%w'.", assignment.FullID(), user, err);
    }

    // Copy the submission to the user's submission directory.
    submissionCopyDir := filepath.Join(submissionDir, GRADING_INPUT_DIRNAME);
    err = util.CopyDirent(submissionPath, submissionCopyDir, true);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to copy submission for assignment '%s' and user '%s': '%w'.", assignment.FullID(), user, err);
    }

    outputDir := filepath.Join(submissionDir, GRADING_OUTPUT_DIRNAME);

    if (options.NoDocker) {
        return RunNoDockerGrader(assignment, submissionPath, outputDir, options);
    }

    return RunDockerGrader(assignment, submissionPath, outputDir, options);
}

func prepSubmissionDir(assignment *model.Assignment, user string, options GradeOptions) (string, error) {
    var submissionDir string;
    var err error;

    if (options.UseFakeSubmissionsDir) {
        tempSubmissionsDir, err := os.MkdirTemp("", "autograding-submissions-dir-");
        if (err != nil) {
            return "", fmt.Errorf("Could not create temp submissions dir: '%w'.", err);
        }

        submissionDir, err = assignment.Course.PrepareSubmissionWithDir(user, tempSubmissionsDir);
        if (err != nil) {
            return "", fmt.Errorf("Failed to prepare fake submission dir: '%w'.", err);
        }

        if (options.LeaveTempDir) {
            log.Info().Str("tempdir", tempSubmissionsDir).Msg("Leaving behind temp submissions dir.");
        } else {
            defer os.RemoveAll(tempSubmissionsDir);
        }
    } else {
        submissionDir, err = assignment.Course.PrepareSubmission(user);
        if (err != nil) {
            return "", fmt.Errorf("Failed to prepare default submission dir: '%w'.", err);
        }
    }

    return submissionDir, nil;
}
