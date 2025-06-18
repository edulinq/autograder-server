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
)

type RegradeOptions struct {
	jobmanager.JobOptions
	GradeOptions

	// The raw references of users to regrade.
	RawReferences []model.CourseUserReference `json:"target-users" required:""`

	// If true, do not swap the context to the background context when running.
	// By default (when this is false), the context will be swapped to the background context when !WaitForCompletion.
	// The swap is so that regrade does not get canceled when an HTTP request is complete.
	// Setting this true is useful for testing (as one round of analysis tests can be wrapped up).
	RetainOriginalContext bool `json:"-"`

	Assignment *model.Assignment `json:"-"`

	ResolvedUsers []string `json:"-"`
}

func Regrade(options RegradeOptions) (map[string]*model.SubmissionHistoryItem, int, map[string]string, error) {
	reference, err := model.ParseCourseUserReferences(options.RawReferences)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("Failed to parse course user references: '%w'.", err)
	}

	courseUsers, err := db.GetCourseUsers(options.Assignment.GetCourse())
	if err != nil {
		return nil, 0, nil, fmt.Errorf("Failed to get course users: '%w'.", err)
	}

	fullUsers := model.ResolveCourseUserEmails(courseUsers, reference)

	if !options.RetainOriginalContext && !options.WaitForCompletion {
		options.Context = context.Background()
	}

	// TODO: We may want to return the proxy time!
	var proxyTime timestamp.Timestamp = 0
	if options.ProxyTime == nil {
		proxyTime = timestamp.Now()
	} else {
		proxyTime = *options.ProxyTime
	}

	lockKey := fmt.Sprintf("regrade-course-%s", options.Assignment.GetCourse().GetID())

	job := jobmanager.Job[string, *model.SubmissionHistoryItem]{
		JobOptions:              &options.JobOptions,
		LockKey:                 lockKey,
		PoolSize:                config.REGRADE_COURSE_POOL_SIZE.Get(),
		ReturnIncompleteResults: !options.WaitForCompletion,
		WorkItems:               fullUsers,
		RetrieveFunc: func(resolvedEmails []string) (map[string]*model.SubmissionHistoryItem, error) {
			return retrieveRegradedSubmissions(options.Assignment, proxyTime, resolvedEmails)
		},
		WorkFunc: func(email string) (*model.SubmissionHistoryItem, error) {
			return computeSingleRegrade(options, email)
		},
		WorkItemKeyFunc: func(email string) string {
			return fmt.Sprintf("%s-%s", lockKey, email)
		},
	}

	err = job.Validate()
	if err != nil {
		return nil, 0, nil, fmt.Errorf("Failed to validate job: '%w'.", err)
	}

	output := job.Run()
	if output.Error != nil {
		return nil, 0, nil, fmt.Errorf("Failed to run regrade job '%s': '%w'.", output.ID, output.Error)
	}

	workErrors := make(map[string]string, len(output.WorkErrors))

	for email, err := range output.WorkErrors {
		workErrors[email] = err.Error()

		logAttributes := make([]any, 3)
		logAttributes = append([]any{err}, log.NewUserAttr(email))
		logAttributes = append(logAttributes, log.NewAttr("job-id", output.ID))
		log.Error("Failed to run regrade.", logAttributes...)
	}

	return output.ResultItems, len(output.RemainingItems), workErrors, nil
}

func computeSingleRegrade(options RegradeOptions, email string) (*model.SubmissionHistoryItem, error) {
	return nil, nil
}

func retrieveRegradedSubmissions(assignment *model.Assignment, proxyTime timestamp.Timestamp, emails []string) (map[string]*model.SubmissionHistoryItem, error) {
	emailSet := make(map[string]any, len(emails))
	for _, email := range emails {
		emailSet[email] = nil
	}

	reference := &model.ParsedCourseUserReference{
		Emails: emailSet,
	}

	results, err := db.GetRecentSubmissionSurvey(assignment, reference)
	if err != nil {
		return nil, fmt.Errorf("Failed to get recent submissions from db: '%w'.", err)
	}

	finalResults := make(map[string]*model.SubmissionHistoryItem, len(results))

	for email, result := range results {
		if result == nil {
			continue
		}

		// The submission was made before the regrade threshold.
		if result.GradingStartTime < proxyTime {
			continue
		}

		finalResults[email] = result
	}

	return finalResults, nil
}
