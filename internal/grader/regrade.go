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
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type RegradeOptions struct {
	Options GradeOptions

	Context    context.Context
	Assignment *model.Assignment

	Users []string
}

func RegradeSubmissions(options RegradeOptions) (map[string]*model.SubmissionHistoryItem, error) {
	if options.Context == nil {
		options.Context = context.Background()
	}

	results, err := runRegradeSubmissions(options)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func runRegradeSubmissions(options RegradeOptions) (map[string]*model.SubmissionHistoryItem, error) {
	if len(options.Users) == 0 {
		return nil, nil
	}

	// Lock based on the course to prevent multiple requests using up all the cores.
	lockKey := fmt.Sprintf("regrade-course-%s", options.Assignment.GetCourse().GetID())
	// TODO: May need noLockWait
	_ = lockmanager.Lock(lockKey)
	defer lockmanager.Unlock(lockKey)

	// The context has been canceled while waiting for a lock, abandon this regrade.
	if options.Context.Err() != nil {
		return nil, nil
	}

	results := make(map[string]*model.SubmissionHistoryItem, len(options.Users))

	poolSize := config.REGRADE_COURSE_POOL_SIZE.Get()
	type PoolResult struct {
		User    string
		Result  *model.SubmissionHistoryItem
		RunTime int64
		// TODO: Should we split into internal vs grading errors?
		Error error
	}

	poolResults, _, err := util.RunParallelPoolMap(poolSize, options.Users, options.Context, func(user string) (PoolResult, error) {
		result, runTime, err := runSingleRegrade(options, user)
		if err != nil {
			err = fmt.Errorf("Failed to perform individual regrade for user %s: '%w'.", user, err)
		}

		return PoolResult{user, result, runTime, err}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to run parallel pool map: '%w'.", err)
	}

	// If the regrade was canceled, exit right away.
	if options.Context.Err() != nil {
		return nil, nil
	}

	var errs error = nil
	totalRunTime := int64(0)

	for _, poolResult := range poolResults {
		if poolResult.Error != nil {
			errs = errors.Join(errs, poolResult.Error)
		} else {
			results[poolResult.User] = poolResult.Result
			totalRunTime += poolResult.RunTime
		}
	}

	// TODO: Look into this!
	// collectIndividualStats(users, totalRunTime, options.ProxyUser)

	return results, errs
}

func runSingleRegrade(options RegradeOptions, user string) (*model.SubmissionHistoryItem, int64, error) {
	previousResult, err := db.GetSubmissionContents(options.Assignment, user, "")
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to get most recent grading result for '%s': '%w'.", user, err)
	}

	if previousResult == nil {
		return nil, 0, nil
	}

	dirName := fmt.Sprintf("regrade-%s-%s-%s-", options.Assignment.GetCourse().GetID(), options.Assignment.GetID(), user)
	tempDir, err := util.MkDirTemp(dirName)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create temp regrade dir: '%w'.", err)
	}
	defer util.RemoveDirent(tempDir)

	err = util.GzipBytesToDirectory(tempDir, previousResult.InputFilesGZip)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to write submission input to a temp dir: '%v'.", err)
	}

	message := ""
	if previousResult.Info != nil {
		message = previousResult.Info.Message
		options.Options.ProxyTime = &previousResult.Info.GradingStartTime
	}

	startTime := timestamp.Now()

	// TODO: Continue handling different failure possibilities.
	// Think about if we can give back partial information.
	// If result.GradingInfo != nil, use ToHistoryItem().
	gradingResult, reject, failureMessage, err := Grade(options.Context, options.Assignment, tempDir, user, message, options.Options)
	if err != nil {
		stdout := ""
		stderr := ""

		if (gradingResult != nil) && (gradingResult.HasTextOutput()) {
			stdout = gradingResult.Stdout
			stderr = gradingResult.Stderr
		}

		log.Warn("Regrade submission failed internally.", err, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr))

		runTime := int64(timestamp.Now() - startTime)
		return nil, runTime, fmt.Errorf("Submission failed internally. stdout: '%s', stderr: '%s': '%w'.", stdout, stderr, err)
	}

	if reject != nil {
		log.Debug("Regrade submission rejected.", log.NewAttr("reason", reject.String()))

		runTime := int64(timestamp.Now() - startTime)
		return nil, runTime, fmt.Errorf("Submission rejected: '%s'.", reject.String())
	}

	if failureMessage != "" {
		log.Debug("Regrade submission got a soft error.", log.NewAttr("message", failureMessage))

		runTime := int64(timestamp.Now() - startTime)
		return nil, runTime, fmt.Errorf("Submission got a soft error: '%s'.", failureMessage)
	}

	var result *model.SubmissionHistoryItem = nil
	if gradingResult.Info != nil {
		result = (*gradingResult.Info).ToHistoryItem()
	}

	runTime := int64(timestamp.Now() - startTime)
	return result, runTime, nil
}
