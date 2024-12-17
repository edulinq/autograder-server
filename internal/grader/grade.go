package grader

import (
	"fmt"
	"sync"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

var submissionLocks sync.Map

type GradeOptions struct {
	NoDocker     bool
	LeaveTempDir bool
    AckLate      bool
}

func GetDefaultGradeOptions(ackLate bool) GradeOptions {
	return GradeOptions{
		NoDocker:     config.DOCKER_DISABLE.Get(),
		LeaveTempDir: config.KEEP_BUILD_DIRS.Get(),
        AckLate:      ackLate,
	}
}

// Grade with default options pulled from config.
func GradeDefault(assignment *model.Assignment, submissionPath string, user string, message string, ackLate bool) (
	*model.GradingResult, RejectReason, string, error) {
	return Grade(assignment, submissionPath, user, message, true, GetDefaultGradeOptions(ackLate))
}

// Grade with custom options.
// Return (result, reject, softGradingError, error).
// Full success is only when ((reject == nil) && (softGradingError == "") && (error == nil)).
func Grade(assignment *model.Assignment, submissionPath string, user string, message string, checkRejection bool, options GradeOptions) (
	*model.GradingResult, RejectReason, string, error) {
	if checkRejection {
		reject, err := checkForRejection(assignment, submissionPath, user, message, options.AckLate)
		if err != nil {
			return nil, nil, "", fmt.Errorf("Failed to check for rejection: '%w'.", err)
		}

		if reject != nil {
			return nil, reject, "", nil
		}
	}

	gradingKey := fmt.Sprintf("%s::%s::%s", assignment.GetCourse().GetID(), assignment.GetID(), user)

	// Get the existing mutex, or store (and fetch) a new one.
	val, _ := submissionLocks.LoadOrStore(gradingKey, &sync.Mutex{})
	lock := val.(*sync.Mutex)

	lock.Lock()
	defer lock.Unlock()

	submissionID, inputFileContents, err := prepForGrading(assignment, submissionPath, user)
	if err != nil {
		return nil, nil, "", fmt.Errorf("Failed to prep for grading: '%w'.", err)
	}

	var gradingResult model.GradingResult
	gradingResult.InputFilesGZip = inputFileContents

	fullSubmissionID := common.CreateFullSubmissionID(assignment.GetCourse().GetID(), assignment.GetID(), user, submissionID)

	var gradingInfo *model.GradingInfo
	var outputFileContents map[string][]byte
	var stdout string
	var stderr string

	startTimestamp := timestamp.Now()

	softGradingError := ""
	if options.NoDocker {
		gradingInfo, outputFileContents, stdout, stderr, err = runNoDockerGrader(assignment, submissionPath, options, fullSubmissionID)
	} else {
		gradingInfo, outputFileContents, stdout, stderr, softGradingError, err = runDockerGrader(assignment, submissionPath, options, fullSubmissionID)
	}

	endTimestamp := timestamp.Now()

	// Copy over stdout and stderr even if an error occured.
	gradingResult.Stdout = stdout
	gradingResult.Stderr = stderr

	// Check for hard grading errors.
	if err != nil {
		return &gradingResult, nil, "", err
	}

	// Check for soft grading errors.
	if softGradingError != "" {
		return &gradingResult, nil, softGradingError, nil
	}

	// Set all the autograder fields in the grading info.
	gradingInfo.ID = fullSubmissionID
	gradingInfo.ShortID = submissionID
	gradingInfo.CourseID = assignment.GetCourse().GetID()
	gradingInfo.AssignmentID = assignment.GetID()
	gradingInfo.User = user
	gradingInfo.Message = message

	if gradingInfo.GradingStartTime.IsZero() {
		gradingInfo.GradingStartTime = startTimestamp
	}

	if gradingInfo.GradingEndTime.IsZero() {
		gradingInfo.GradingEndTime = endTimestamp
	}

	gradingInfo.ComputePoints()

	gradingResult.Info = gradingInfo
	gradingResult.OutputFilesGZip = outputFileContents

	err = db.SaveSubmission(assignment, &gradingResult)
	if err != nil {
		return &gradingResult, nil, "", fmt.Errorf("Failed to save grading result: '%w'.", err)
	}

	return &gradingResult, nil, "", nil
}

func prepForGrading(assignment *model.Assignment, submissionPath string, user string) (string, map[string][]byte, error) {
	// Ensure the assignment docker image is built.
	err := docker.BuildImageFromSourceQuick(assignment)
	if err != nil {
		return "", nil, fmt.Errorf("Failed to build assignment '%s' docker image: '%w'.", assignment.FullID(), err)
	}

	submissionID, err := db.GetNextSubmissionID(assignment, user)
	if err != nil {
		return "", nil, fmt.Errorf("Unable to get next submission id for assignment '%s', user '%s': '%w'.", assignment.FullID(), user, err)
	}

	fileContents, err := util.GzipDirectoryToBytes(submissionPath)
	if err != nil {
		return "", nil, fmt.Errorf("Failed to copy submission input '%s': '%w'.", submissionPath, err)
	}

	return submissionID, fileContents, nil
}
