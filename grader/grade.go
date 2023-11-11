package grader

import (
    "fmt"
    "sync"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var submissionLocks sync.Map;

type GradeOptions struct {
    NoDocker bool
    LeaveTempDir bool
}

func GetDefaultGradeOptions() GradeOptions {
    return GradeOptions{
        NoDocker: config.DOCKER_DISABLE.Get(),
        LeaveTempDir: config.DEBUG.Get(),
    };
}

// Grade with default options pulled from config.
func GradeDefault(assignment model.Assignment, submissionPath string, user string, message string) (*artifact.GradingResult, error) {
    return Grade(assignment, submissionPath, user, message, GetDefaultGradeOptions());
}

// Grade with custom options.
func Grade(assignment model.Assignment, submissionPath string, user string, message string, options GradeOptions) (*artifact.GradingResult, error) {
    var gradingResult artifact.GradingResult;

    gradingKey := fmt.Sprintf("%s::%s::%s", assignment.GetCourse().GetID(), assignment.GetID(), user);

    // Get the existing mutex, or store (and fetch) a new one.
    val, _ := submissionLocks.LoadOrStore(gradingKey, &sync.Mutex{});
    lock := val.(*sync.Mutex)

    lock.Lock();
    defer lock.Unlock()

    submissionID, inputFileContents, err := prepForGrading(assignment, submissionPath, user);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to prep for grading: '%w'.", err);
    }

    gradingResult.InputFilesGZip = inputFileContents;

    fullSubmissionID := common.CreateFullSubmissionID(assignment.GetCourse().GetID(), assignment.GetID(), user, submissionID);

    var submissionResult *artifact.GradedAssignment;
    var outputFileContents map[string][]byte;
    var stdout string;
    var stderr string;

    if (options.NoDocker) {
        submissionResult, outputFileContents, stdout, stderr, err = runNoDockerGrader(assignment, submissionPath, options, fullSubmissionID);
    } else {
        submissionResult, outputFileContents, stdout, stderr, err = runDockerGrader(assignment, submissionPath, options, fullSubmissionID);
    }

    // Copy over stdout and stderr even if an error occured.
    gradingResult.Stdout = stdout;
    gradingResult.Stderr = stderr;

    if (err != nil) {
        return &gradingResult, err;
    }

    // Set all the autograder fields in the submissionResult.
    submissionResult.ID = fullSubmissionID;
    submissionResult.ShortID = submissionID;
    submissionResult.CourseID = assignment.GetCourse().GetID();
    submissionResult.AssignmentID = assignment.GetID();
    submissionResult.User = user;
    submissionResult.Message = message;
    submissionResult.ComputePoints();

    gradingResult.Result = submissionResult;
    gradingResult.OutputFilesGZip = outputFileContents;

    if (!config.NO_STORE.Get()) {
        err = db.SaveSubmission(assignment, &gradingResult);
        if (err != nil) {
            return &gradingResult, fmt.Errorf("Failed to save grading result: '%w'.", err);
        }
    }

    return &gradingResult, nil;
}

func prepForGrading(assignment model.Assignment, submissionPath string, user string) (string, map[string][]byte, error) {
    // Ensure the assignment docker image is built.
    err := assignment.BuildImageQuick();
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to build assignment assignment '%s' docker image: '%w'.", assignment.FullID(), err);
    }

    submissionID, err := db.GetNextSubmissionID(assignment, user);
    if (err != nil) {
        return "", nil, fmt.Errorf("Unable to get next submission id for assignment '%s', user '%s': '%w'.", assignment.FullID(), user, err);
    }

    fileContents, err := util.GzipDirectoryToBytes(submissionPath);
    if (err != nil) {
        return "", nil, fmt.Errorf("Failed to copy submission input '%s': '%w'.", submissionPath, err);
    }

    return submissionID, fileContents, nil;
}
