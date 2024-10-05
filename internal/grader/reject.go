package grader

import (
	"fmt"
	"time"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
)

// Reasons a submission can be rejected.
type RejectReason interface {
	String() string
}

type RejectMaxAttempts struct {
	Max int
}

func (this *RejectMaxAttempts) String() string {
	return fmt.Sprintf("Reached the number of max attempts: %d.", this.Max)
}

type RejectWindowMax struct {
	Max                int
	WindowDuration     common.DurationSpec
	EarliestSubmission timestamp.Timestamp
}

func (this *RejectWindowMax) String() string {
	return this.fullString(timestamp.Now())
}

func (this *RejectWindowMax) fullString(now timestamp.Timestamp) string {
	nextTime := timestamp.FromMSecs(this.EarliestSubmission.ToMSecs() + this.WindowDuration.TotalMSecs())
	deltaMS := nextTime.ToMSecs() - now.ToMSecs()
	deltaString := time.Duration(deltaMS * int64(time.Millisecond)).String()

	return fmt.Sprintf("Reached the number of max attempts (%d) within submission window (%s)."+
		" Next allowed submission time is %s (in %s).",
		this.Max, this.WindowDuration.ShortString(),
		nextTime.SafeMessage(), deltaString)
}

func checkForRejection(assignment *model.Assignment, submissionPath string, user string, message string) (RejectReason, error) {
	return checkSubmissionLimit(assignment, user)
}

func checkSubmissionLimit(assignment *model.Assignment, email string) (RejectReason, error) {
	// Do not check for submission limits in testing mode.
	if config.UNIT_TESTING_MODE.Get() {
		return nil, nil
	}

	user, err := db.GetCourseUser(assignment.GetCourse(), email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Unable to find user: '%s'.", email)
	}

	// User that are >= grader are not subject to submission restrictions.
	if user.Role >= model.CourseRoleGrader {
		return nil, nil
	}

	limit := assignment.GetSubmissionLimit()
	if limit == nil {
		return nil, nil
	}

	now := timestamp.Now()

	history, err := db.GetSubmissionHistory(assignment, email)
	if err != nil {
		return nil, err
	}

	if *limit.Max >= 0 {
		if len(history) >= *limit.Max {
			return &RejectMaxAttempts{*limit.Max}, nil
		}
	}

	if limit.Window != nil {
		reason, err := checkSubmissionLimitWindow(limit.Window, history, now)
		if err != nil {
			return nil, err
		}

		if reason != nil {
			return reason, nil
		}
	}

	return nil, nil
}

func checkSubmissionLimitWindow(window *model.SubmittionLimitWindow,
	history []*model.SubmissionHistoryItem, now timestamp.Timestamp) (RejectReason, error) {
	if len(history) < window.AllowedAttempts {
		return nil, nil
	}

	windowStart := timestamp.FromMSecs(now.ToMSecs() - window.Duration.TotalMSecs())
	earliestTime := timestamp.Zero()

	windowCount := 0
	for _, item := range history {
		if item.GradingStartTime > windowStart {
			windowCount++

			if earliestTime.IsZero() || (earliestTime > item.GradingStartTime) {
				earliestTime = item.GradingStartTime
			}
		}
	}

	if windowCount >= window.AllowedAttempts {
		return &RejectWindowMax{window.AllowedAttempts, window.Duration, earliestTime}, nil
	}

	return nil, nil
}
