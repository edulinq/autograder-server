package grader

import (
	"fmt"
	"time"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
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
	WindowDuration     util.DurationSpec
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

type RejectLate struct {
	AssignmentName string
	DueDate        timestamp.Timestamp
}

func (this *RejectLate) String() string {
	deltaMS := timestamp.Now().ToMSecs() - this.DueDate.ToMSecs()
	deltaString := time.Duration(deltaMS * int64(time.Millisecond)).String()

	return fmt.Sprintf("Attempting to submit assignment (%s) late without the 'allow late' option."+
		" It was due on %s (which was %s ago)."+
		" Use the 'allow late' option to submit an assignment late (e.g., `--allow-late`)."+
		" See your interface's documentation for more information.",
		this.AssignmentName, this.DueDate.SafeMessage(), deltaString)
}

func checkForRejection(assignment *model.Assignment, submissionPath string, email string, message string, allowLate bool) (RejectReason, error) {
	user, err := db.GetServerUser(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Unable to find user: '%s'.", email)
	}

	// Server admins are never rejected.
	if user.Role >= model.ServerRoleAdmin {
		return nil, nil
	}

	reason := checkLateSubmission(assignment, allowLate)
	if reason != nil {
		return reason, nil
	}

	return checkSubmissionLimit(assignment, email)
}

func checkLateSubmission(assignment *model.Assignment, allowLate bool) RejectReason {
	if assignment.DueDate == nil {
		return nil
	}

	now := timestamp.Now()

	if (now > *assignment.DueDate) && !allowLate {
		return &RejectLate{assignment.Name, *assignment.DueDate}
	}

	return nil
}

func checkSubmissionLimit(assignment *model.Assignment, email string) (RejectReason, error) {
	// Do not check for submission limits in testing mode.
	if config.UNIT_TESTING_MODE.Get() {
		return nil, nil
	}

	// Note that server admins were already checked for in checkForRejection(),
	// so we don't need to worry about escalation here.
	user, err := db.GetCourseUser(assignment.GetCourse(), email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Unable to find user: '%s'.", email)
	}

	// Users that are >= grader are not subject to submission restrictions.
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
		if countStudentSubmissions(history) >= *limit.Max {
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

// Count student-initiated submissions, excluding proxy submissions.
func countStudentSubmissions(history []*model.SubmissionHistoryItem) int {
	count := 0
	for _, item := range history {
		// Skip proxy submissions - they don't count against student limits.
		if item.ProxyUser != "" {
			continue
		}

		count++
	}

	return count
}

func checkSubmissionLimitWindow(window *model.SubmittionLimitWindow,
	history []*model.SubmissionHistoryItem, now timestamp.Timestamp) (RejectReason, error) {
	if countStudentSubmissions(history) < window.AllowedAttempts {
		return nil, nil
	}

	windowStart := timestamp.FromMSecs(now.ToMSecs() - window.Duration.TotalMSecs())
	earliestTime := timestamp.Zero()

	windowCount := 0
	for _, item := range history {
		// Skip proxy submissions - they don't count against student limits.
		if item.ProxyUser != "" {
			continue
		}

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
