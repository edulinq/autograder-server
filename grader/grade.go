package grader

import (
    "fmt"
    "os"
    "path/filepath"
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
func GradeDefault(assignment model.Assignment, submissionPath string, user string, message string) (*artifact.GradedAssignment, *artifact.SubmissionSummary, string, error) {
    return Grade(assignment, submissionPath, user, message, GetDefaultGradeOptions());
}

// Grade with custom options.
func Grade(assignment model.Assignment, submissionPath string, user string, message string, options GradeOptions) (*artifact.GradedAssignment, *artifact.SubmissionSummary, string, error) {
    gradingKey := fmt.Sprintf("%s::%s::%s", assignment.GetCourse().GetID(), assignment.GetID(), user);

    // Get the existing mutex, or store (and fetch) a new one.
    val, _ := submissionLocks.LoadOrStore(gradingKey, &sync.Mutex{});
    lock := val.(*sync.Mutex)

    lock.Lock();
    defer lock.Unlock();

    // Ensure the assignment docker image is built.
    err := assignment.BuildImageQuick();
    if (err != nil) {
        return nil, nil, "", fmt.Errorf("Failed to build assignment assignment '%s' docker image: '%w'.", assignment.FullID(), err);
    }

    submissionID, err := db.GetNextSubmissionID(assignment, user);
    if (err != nil) {
        return nil, nil, "", fmt.Errorf("Unable to get next submission id for assignment'%s', user '%s': '%w'.", assignment.FullID(), user, err);
    }

    submissionDir, err := util.MkDirTemp("autograder-submission-dir-");
    if (err != nil) {
        return nil, nil, "", fmt.Errorf("Could not create temp submission dir: '%w'.", err);
    }


    fullSubmissionID := common.CreateFullSubmissionID(assignment.GetCourse().GetID(), assignment.GetID(), user, submissionID);

    // Copy the submission to the user's submission directory.
    submissionCopyDir := filepath.Join(submissionDir, common.GRADING_INPUT_DIRNAME);
    err = util.CopyDirent(submissionPath, submissionCopyDir, true);
    if (err != nil) {
        return nil, nil, "", fmt.Errorf("Failed to copy submission for assignment '%s' and user '%s': '%w'.", assignment.FullID(), user, err);
    }

    outputDir := filepath.Join(submissionDir, common.GRADING_OUTPUT_DIRNAME);
    os.MkdirAll(outputDir, 0755);

    var result *artifact.GradedAssignment;
    var output string;

    if (options.NoDocker) {
        result, output, err = RunNoDockerGrader(assignment, submissionPath, outputDir, options, fullSubmissionID);
    } else {
        result, output, err = RunDockerGrader(assignment, submissionPath, outputDir, options, fullSubmissionID);
    }

    if (err != nil) {
        return nil, nil, output, err;
    }

    // Set all the autograder fields in the result.
    result.ID = fullSubmissionID;
    result.ShortID = submissionID;
    result.CourseID = assignment.GetCourse().GetID();
    result.AssignmentID = assignment.GetID();
    result.User = user;
    result.Message = message;
    result.ComputePoints();

    // TEST - Remove summary
    summary := result.GetSummary(fullSubmissionID, message);
    summaryPath := filepath.Join(outputDir, common.GRADER_OUTPUT_SUMMARY_FILENAME);

    err = util.ToJSONFileIndent(summary, summaryPath);
    if (err != nil) {
        return nil, nil, output, fmt.Errorf("Failed to write submission summary for assignment '%s' and user '%s': '%w'.", assignment.FullID(), user, err);
    }

    // TEST - Save results in DB?

    // TEST - Skip DB save if config.NO_STORE ?

    return result, summary, output, nil;
}
