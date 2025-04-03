package grader

import (
	"context"
	"fmt"
	"time"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// Allow for a little extra runtime for setup/cleaup when running the grader.
// This extra time is just for the safety context around the actual graders (which have their own timeouts).
const extraRunTimeSecs int = 10

type GradeOptions struct {
	NoDocker       bool
	LeaveTempDir   bool
	CheckRejection bool
	AllowLate      bool
	ProxyUser      string
	ProxyTime      *timestamp.Timestamp
}

func GetDefaultGradeOptions() GradeOptions {
	return GradeOptions{
		NoDocker:       config.DOCKER_DISABLE.Get(),
		LeaveTempDir:   config.KEEP_BUILD_DIRS.Get(),
		CheckRejection: true,
		AllowLate:      false,
		ProxyUser:      "",
		ProxyTime:      nil,
	}
}

// Grade with default options pulled from config.
func GradeDefault(assignment *model.Assignment, submissionPath string, user string, message string) (
	*model.GradingResult, RejectReason, string, error) {
	return Grade(context.Background(), assignment, submissionPath, user, message, GetDefaultGradeOptions())
}

// Grade with custom options.
// Return (result, reject, softGradingError, error).
// Full success is only when ((reject == nil) && (softGradingError == "") && (error == nil)).
func Grade(ctx context.Context, assignment *model.Assignment, submissionPath string, user string, message string, options GradeOptions) (
	*model.GradingResult, RejectReason, string, error) {
	if options.CheckRejection {
		reject, err := checkForRejection(assignment, submissionPath, user, message, options.AllowLate)
		if err != nil {
			return nil, nil, "", fmt.Errorf("Failed to check for rejection: '%w'.", err)
		}

		if reject != nil {
			return nil, reject, "", nil
		}
	}

	gradingKey := fmt.Sprintf("%s::%s::%s", assignment.GetCourse().GetID(), assignment.GetID(), user)

	// Get the grading start time right before we acquire the user's lock.
	startTimestamp := timestamp.Now()

	// Ensure the user can only have one submission (of each assignment) running at a time.
	lockmanager.Lock(gradingKey)
	defer lockmanager.Unlock(gradingKey)

	submissionID, inputFileContents, err := prepForGrading(assignment, submissionPath, user)
	if err != nil {
		return nil, nil, "", fmt.Errorf("Failed to prep for grading: '%w'.", err)
	}

	var gradingResult model.GradingResult
	gradingResult.InputFilesGZip = inputFileContents

	fullSubmissionID := common.CreateFullSubmissionID(assignment.GetCourse().GetID(), assignment.GetID(), user, submissionID)

	gradingInfo, outputFileContents, stdout, stderr, softGradingError, err := runGrader(ctx, assignment, submissionPath, options, fullSubmissionID)

	endTimestamp := timestamp.Now()

	// Copy over stdout and stderr even if an error occurred.
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
	gradingInfo.ProxyUser = options.ProxyUser

	if options.ProxyTime == nil {
		gradingInfo.GradingStartTime = startTimestamp
		gradingInfo.GradingEndTime = endTimestamp
	} else {
		gradingInfo.GradingStartTime = *options.ProxyTime
		gradingInfo.GradingEndTime = *options.ProxyTime + (endTimestamp - startTimestamp)
		gradingInfo.ProxyStartTime = &startTimestamp
		gradingInfo.ProxyEndTime = &endTimestamp
	}

	gradingInfo.ComputePoints()

	gradingResult.Info = gradingInfo
	gradingResult.OutputFilesGZip = outputFileContents

	err = db.SaveSubmission(assignment, &gradingResult)
	if err != nil {
		return &gradingResult, nil, "", fmt.Errorf("Failed to save grading result: '%w'.", err)
	}

	metric := stats.Metric{
		Timestamp: startTimestamp,
		Type:      stats.MetricTypeGradingTime,
		Value:     float64((endTimestamp - startTimestamp).ToMSecs()),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeUserEmail:    gradingInfo.User,
			stats.MetricAttributeCourseID:     gradingInfo.CourseID,
			stats.MetricAttributeAssignmentID: gradingInfo.AssignmentID,
		},
	}

	// Store stats for this grading (when everything is successful).
	stats.AsyncStoreMetric(&metric)

	return &gradingResult, nil, "", nil
}

// Resolve a missing proxy time for a given assignment.
// If the submission is not late (including no due date), return the current time.
// Otherwise, the proxy time will be set to one minute before the due date.
func ResolveProxyTime(proxyTime *timestamp.Timestamp, assignment *model.Assignment) *timestamp.Timestamp {
	if proxyTime != nil {
		return proxyTime
	}

	now := timestamp.Now()
	if assignment.DueDate == nil {
		return &now
	}

	if now < *assignment.DueDate {
		return &now
	}

	oneMinute := time.Duration(1 * time.Minute).Milliseconds()
	minuteBeforeDueDate := *assignment.DueDate - timestamp.FromMSecs(oneMinute)

	return &minuteBeforeDueDate
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

func getTimeoutMessage(assignment *model.Assignment) string {
	return fmt.Sprintf("Submission has ran for too long and was killed. Max assignment runtime is %d seconds (server hard limit is %d seconds). Check for infinite loops/recursion and consult with your instructors/TAs.", assignment.MaxRuntimeSecs, config.GRADING_RUNTIME_MAX_SECS.Get())
}

func getCanceledMessage(assignment *model.Assignment) string {
	return "Grading has been canceled (usually by a broken HTTP connection)."
}

// Add an additional level for waiting for timeouts.
// Timeouts should be handled a level below this (e.g., docker or exec),
// but this is an additional layer just in case there are issues at that level.
func runGrader(ctx context.Context, assignment *model.Assignment, submissionPath string, options GradeOptions, fullSubmissionID string) (*model.GradingInfo, map[string][]byte, string, string, string, error) {
	var gradingInfo *model.GradingInfo
	var outputFileContents map[string][]byte
	var stdout string
	var stderr string
	var softGradingError string
	var err error

	runFunc := func() {
		if options.NoDocker {
			gradingInfo, outputFileContents, stdout, stderr, softGradingError, err = runNoDockerGrader(ctx, assignment, submissionPath, options, fullSubmissionID)
		} else {
			gradingInfo, outputFileContents, stdout, stderr, softGradingError, err = runDockerGrader(ctx, assignment, submissionPath, options, fullSubmissionID)
		}
	}

	timeoutMS := int64((assignment.MaxRuntimeSecs + extraRunTimeSecs) * 1000)
	ok := util.RunWithTimeout(timeoutMS, runFunc)

	if !ok {
		// Timeout
		// We must return very general results (which is why we prefer to catch this at the grader level).
		return nil, nil, "", "", getTimeoutMessage(assignment), nil
	}

	return gradingInfo, outputFileContents, stdout, stderr, softGradingError, err
}
