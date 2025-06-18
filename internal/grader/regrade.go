package grader

import (
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/jobmanager"
)

type RegradeOptions struct {
	jobmanager.JobOptions
	GradeOptions

	// The raw references of users to regrade.
	RawReferences []CourseUserReferences `json:"target-users" required:""`

	// If true, do not swap the context to the background context when running.
	// By default (when this is false), the context will be swapped to the background context when !WaitForCompletion.
	// The swap is so that regrade does not get canceled when an HTTP request is complete.
	// Setting this true is useful for testing (as one round of analysis tests can be wrapped up).
	RetainOriginalContext bool `json:"-"`

	Assignment *model.Assignment `json:"-"`

	ResolvedUsers []string `json:"-"`
}

func Regrade(options RegradeOptions) (map[string]*model.SubmissionHistoryItem, int, map[string]string, error) {
	reference, err := ParseCourseUserReferences(options.RawReferences)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("Failed to parse course user references: '%w'.", err)
	}

	fullUsers := model.ResolveCourseUserEmails(db.GetCourseUsers(options.Assignment.GetCourse().GetID(), reference))

	if !options.RetainOriginalContext && !options.WaitForCompletion {
		options.Context = context.Background()
	}

	job := jobmanager.Job[string, *model.IndividualAnalysis]{
		JobOptions:              &options.JobOptions,
		LockKey:                 fmt.Sprintf("regrade-course-%s", options.Assignment.GetCourse().GetID()),
		PoolSize:                config.REGRADE_COURSE_POOL_SIZE.Get(),
		ReturnIncompleteResults: !options.WaitForCompletion,
		WorkItems:               fullUsers,
		// TODO: Does this need to be modified to not include nil results? Guess: probably
		RetrieveFunc: db.GetRecentSubmissionSurvey(options.Assignment, reference),
		StoreFunc:    db.StoreIndividualAnalysis,
		RemoveFunc:   db.RemoveIndividualAnalysis,
		WorkFunc: func(fullSubmissionID string) (*model.IndividualAnalysis, error) {
			return computeSingleIndividualAnalysis(options, fullSubmissionID, true)
		},
		WorkItemKeyFunc: func(fullSubmissionID string) string {
			return fmt.Sprintf("analysis-individual-%s", fullSubmissionID)
		},
		OnComplete: func(result *jobmanager.JobOutput[string, *model.IndividualAnalysis]) {
			if result == nil {
				return
			}

			collectIndividualStats(fullSubmissionIDs, result.RunTime, options.InitiatorEmail)
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
