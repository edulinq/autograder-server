package grader

import (
	"context"
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

type RegradeOptions struct {
	Users []string `json:"users"`

	// Wait for the entire regrade to complete and return all results.
	WaitForCompletion bool `json:"wait-for-completion"`

	Options GradeOptions `json:"-"`

	// A context that can be used to cancel the regrade.
	Context    context.Context   `json:"-"`
	Assignment *model.Assignment `json:"-"`

	// If true, do not swap the context to the background context when running.
	// By default (when this is false), the context will be swapped to the background context when !WaitForCompletion.
	// The swap is so that regrade does not get canceled when an HTTP request is complete.
	// Setting this true is useful for testing (as one round of regrade tests can be wrapped up).
	RetainOriginalContext bool `json:"-"`
}

func RegradeSubmissions(options RegradeOptions) (map[string]*model.SubmissionHistoryItem, int, error) {
	if options.Context == nil {
		options.Context = context.Background()
	}

	if options.WaitForCompletion {
		results, err := runRegradeSubmissions(options)
		if err != nil {
			return nil, 0, err
		}

		return results, 0, nil
	} else {
		if !options.RetainOriginalContext {
			options.Context = context.Background()
		}

		go func() {
			_, err := runRegradeSubmissions(options)
			if err != nil {
				log.Error("Failure during asynchronous regrading.", err)
			}
		}()
	}

	return map[string]*model.SubmissionHistoryItem{}, len(options.Users), nil
}

func runRegradeSubmissions(options RegradeOptions) (map[string]*model.SubmissionHistoryItem, error) {
	if len(options.Users) == 0 {
		return nil, nil
	}

	// Lock based on the course to prevent multiple requests using up all the cores.
	lockKey := fmt.Sprintf("regrade-course-%s", options.Assignment.GetCourse().GetID())
	lockmanager.Lock(lockKey)
	defer lockmanager.Unlock(lockKey)

	// The context has been canceled while waiting for a lock, abandon this regrade.
	if options.Context.Err() != nil {
		return nil, nil
	}

	results := make(map[string]*model.SubmissionHistoryItem, len(options.Users))

	poolSize := config.REGRADE_COURSE_POOL_SIZE.Get()
	type PoolResult struct {
		User   string
		Result *model.SubmissionHistoryItem
		Error  error
	}

	poolResults, _, err := util.RunParallelPoolMap(poolSize, options.Users, options.Context, func(user string) (PoolResult, error) {
		result, err := runSingleRegrade(options, user)
		if err != nil {
			err = fmt.Errorf("Failed to perform individual regrade for user %s: '%w'.", user, err)
		}

		return PoolResult{user, result, err}, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to run parallel pool map: '%w'.", err)
	}

	// If the regrade was canceled, exit right away.
	if options.Context.Err() != nil {
		return nil, nil
	}

	var errs error = nil

	for _, poolResult := range poolResults {
		if poolResult.Error != nil {
			errs = errors.Join(errs, poolResult.Error)
		} else {
			results[poolResult.User] = poolResult.Result
		}
	}

	return results, errs
}

func runSingleRegrade(options RegradeOptions, user string) (*model.SubmissionHistoryItem, error) {
	previousResult, err := db.GetSubmissionContents(options.Assignment, user, "")
	if err != nil {
		return nil, fmt.Errorf("Failed to get most recent grading result for '%s': '%w'.", user, err)
	}

	if previousResult == nil {
		return nil, nil
	}

	dirName := fmt.Sprintf("regrade-%s-%s-%s-", options.Assignment.GetCourse().GetID(), options.Assignment.GetID(), user)
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
		options.Options.ProxyTime = &previousResult.Info.GradingStartTime
	}

	gradingResult, reject, failureMessage, err := Grade(options.Context, options.Assignment, tempDir, user, message, options.Options)
	if err != nil {
		stdout := ""
		stderr := ""

		if (gradingResult != nil) && (gradingResult.HasTextOutput()) {
			stdout = gradingResult.Stdout
			stderr = gradingResult.Stderr
		}

		log.Warn("Regrade submission failed internally.", err, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr))

		return nil, nil
	}

	if reject != nil {
		log.Debug("Regrade submission rejected.", log.NewAttr("reason", reject.String()))

		return nil, nil
	}

	if failureMessage != "" {
		log.Debug("Regrade submission got a soft error.", log.NewAttr("message", failureMessage))

		return nil, nil
	}

	var result *model.SubmissionHistoryItem = nil
	if gradingResult.Info != nil {
		result = (*gradingResult.Info).ToHistoryItem()
	}

	return result, nil
}
