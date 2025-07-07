package grader

import (
	"context"
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type RegradeOptions struct {
	jobmanager.JobOptions
	GradeOptions `json:"-"`

	// The raw references of users to regrade.
	RawReferences []model.CourseUserReference `json:"target-users" required:""`

	// Ensure every user has made a new submission after this time.
	// If nil, the current time will be used.
	RegradeAfter *timestamp.Timestamp `json:"regrade-after"`

	// If true, do not swap the context to the background context when running.
	// By default (when this is false), the context will be swapped to the background context when !WaitForCompletion.
	// The swap is so that regrade does not get canceled when an HTTP request is complete.
	// Setting this true is useful for testing (as one round of analysis tests can be wrapped up).
	RetainOriginalContext bool `json:"-"`

	ResolvedUsers []string `json:"-"`
}

type RegradeResult struct {
	Options      RegradeOptions                          `json:"options"`
	RegradeAfter timestamp.Timestamp                     `json:"regrade-after"`
	Results      map[string]*model.SubmissionHistoryItem `json:"results"`
	WorkErrors   map[string]string                       `json:"work-errors"`
}

func Regrade(assignment *model.Assignment, options RegradeOptions) (*RegradeResult, int, error) {
	reference, err := model.ParseCourseUserReferences(options.RawReferences)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to parse course user references: '%w'.", err)
	}

	courseUsers, err := db.GetCourseUsers(assignment.GetCourse())
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to get course users: '%w'.", err)
	}

	options.ResolvedUsers = model.ResolveCourseUserEmails(courseUsers, reference)

	if !options.RetainOriginalContext && !options.WaitForCompletion {
		options.Context = context.Background()
	}

	var regradeAfter timestamp.Timestamp = 0
	if options.RegradeAfter == nil {
		regradeAfter = timestamp.Now()
	} else {
		regradeAfter = *options.RegradeAfter
	}

	lockKey := fmt.Sprintf("regrade-course-%s", assignment.GetCourse().GetID())

	job := jobmanager.Job[string, *model.SubmissionHistoryItem]{
		JobOptions:              &options.JobOptions,
		LockKey:                 lockKey,
		PoolSize:                config.REGRADE_COURSE_POOL_SIZE.Get(),
		ReturnIncompleteResults: !options.WaitForCompletion,
		WorkItems:               options.ResolvedUsers,
		RetrieveFunc: func(resolvedEmails []string) (map[string]*model.SubmissionHistoryItem, error) {
			return retrieveRegradedSubmissions(assignment, regradeAfter, resolvedEmails)
		},
		WorkFunc: func(email string) (*model.SubmissionHistoryItem, error) {
			return computeSingleRegrade(courseUsers, assignment, options, email)
		},
		WorkItemKeyFunc: func(email string) string {
			return fmt.Sprintf("%s-%s", lockKey, email)
		},
	}

	err = job.Validate()
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to validate job: '%w'.", err)
	}

	output := job.Run()
	if output.Error != nil {
		return nil, 0, fmt.Errorf("Failed to run regrade job '%s': '%w'.", output.ID, output.Error)
	}

	workErrors := make(map[string]string, len(output.WorkErrors))

	for email, err := range output.WorkErrors {
		workErrors[email] = err.Error()

		logAttributes := make([]any, 3)
		logAttributes = append([]any{err}, log.NewUserAttr(email))
		logAttributes = append(logAttributes, log.NewAttr("job-id", output.ID))
		log.Error("Failed to run regrade.", logAttributes...)
	}

	regradeResult := RegradeResult{
		Options:      options,
		RegradeAfter: regradeAfter,
		Results:      output.ResultItems,
		WorkErrors:   workErrors,
	}

	return &regradeResult, len(output.RemainingItems), nil
}

func computeSingleRegrade(courseUsers map[string]*model.CourseUser, assignment *model.Assignment, options RegradeOptions, email string) (*model.SubmissionHistoryItem, error) {
	_, ok := courseUsers[email]
	if !ok {
		return nil, fmt.Errorf("Cannot regrade an unknown user: '%s'.", email)
	}

	previousResult, err := db.GetSubmissionContents(assignment, email, "")
	if err != nil {
		return nil, fmt.Errorf("Failed to get most recent grading result for '%s': '%w'.", email, err)
	}

	if previousResult == nil {
		return nil, nil
	}

	dirName := fmt.Sprintf("regrade-%s-%s-%s-", assignment.GetCourse().GetID(), assignment.GetID(), email)
	tempDir, err := util.MkDirTemp(dirName)
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp regrade dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	err = util.GzipBytesToDirectory(tempDir, previousResult.InputFilesGZip)
	if err != nil {
		return nil, fmt.Errorf("Failed to write submission input to a temp dir: '%v'.", err)
	}

	message := ""
	if previousResult.Info != nil {
		message = previousResult.Info.Message
		options.GradeOptions.ProxyTime = &previousResult.Info.GradingStartTime
	}

	gradingResult, reject, failureMessage, err := Grade(options.Context, assignment, tempDir, email, message, options.GradeOptions)
	if err != nil {
		stdout := ""
		stderr := ""

		if (gradingResult != nil) && (gradingResult.HasTextOutput()) {
			stdout = gradingResult.Stdout
			stderr = gradingResult.Stderr
		}

		return nil, fmt.Errorf("Regrade submission failed internally: '%w'. Stdout: '%s', Stderr: '%s'.", err, stdout, stderr)
	}

	if reject != nil {
		return nil, fmt.Errorf("Regrade submission rejected: '%s'.", reject.String())
	}

	if failureMessage != "" {
		return nil, fmt.Errorf("Regrade submission got a soft error: '%s'.", failureMessage)
	}

	var result *model.SubmissionHistoryItem = nil
	if gradingResult.Info != nil {
		result = (*gradingResult.Info).ToHistoryItem()
	}

	return result, nil
}

func retrieveRegradedSubmissions(assignment *model.Assignment, regradeAfter timestamp.Timestamp, emails []string) (map[string]*model.SubmissionHistoryItem, error) {
	emailSet := make(map[string]any, len(emails))
	for _, email := range emails {
		emailSet[email] = nil
	}

	reference := &model.ParsedCourseUserReference{
		Emails: emailSet,
	}

	results, err := db.GetRecentSubmissions(assignment, reference)
	if err != nil {
		return nil, fmt.Errorf("Failed to get recent submissions from db: '%w'.", err)
	}

	finalResults := make(map[string]*model.SubmissionHistoryItem, len(results))

	for email, result := range results {
		if result == nil {
			continue
		}

		// The submission was made before the regrade threshold.
		if isSubmittedBeforeRegradeTime(result.GradingStartTime, result.ProxyStartTime, regradeAfter) {
			continue
		}

		finalResults[email] = result.ToHistoryItem()
	}

	return finalResults, nil
}

func isSubmittedBeforeRegradeTime(gradingStartTime timestamp.Timestamp, proxyStartTime *timestamp.Timestamp, regradeAfter timestamp.Timestamp) bool {
	if gradingStartTime >= regradeAfter {
		return false
	}

	// Grading start time is before the regrade after time.
	// Check for proxy submissions made after the threshold.

	if proxyStartTime == nil {
		// The submission is not a proxy submission.
		return true
	}

	return *proxyStartTime < regradeAfter
}
