package grader

import (
    "fmt"
    "time"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/model"
)

// Reasons a submission can be rejected.
type RejectReason interface {
    String() string
}

type RejectMaxAttempts struct {
    Max int
}

func (this *RejectMaxAttempts) String() string {
    return fmt.Sprintf("Reached the number of max attempts: %d.", this.Max);
}

type RejectWindowMax struct {
    Max int
    WindowDuration common.DurationSpec
    EarliestSubmission time.Time
}

func (this *RejectWindowMax) String() string {
    nextTime := this.EarliestSubmission.Add(time.Duration(this.WindowDuration.TotalNanosecs()));
    delta := nextTime.Sub(time.Now());
    return fmt.Sprintf("Reached the number of max attempts (%d) within submission window (%s)." +
            " Next allowed submission time is %s (in %s).",
            this.Max, this.WindowDuration.ShortString(),
            nextTime.Format(time.RFC1123), delta.String());
}

type RejectMissingLateAcknowledgment struct {}

func (this *RejectMissingLateAcknowledgment) String() string {
    return "This assignment is past the due date and the late policy will be applied."
}

func checkForRejection(assignment *model.Assignment, submissionPath string, user string, message string, lateAcknowledgment bool) (RejectReason, error) {
    reject := checkLateAcknowledgment(assignment, lateAcknowledgment)
    if reject != nil {
        return reject, nil
    }
    return checkSubmissionLimit(assignment, user);
}

func checkSubmissionLimit(assignment *model.Assignment, email string) (RejectReason, error) {
    // Do not check for submission limits in testing mode.
    if (config.TESTING_MODE.Get()) {
        return nil, nil;
    }

    user, err := db.GetUser(assignment.GetCourse(), email);
    if (err != nil) {
        return nil, err;
    }

    if (user == nil) {
        return nil, fmt.Errorf("Unable to find user: '%s'.", email);
    }

    // User that are >= grader are not subject to submission restrictions.
    if (user.Role >= model.RoleGrader) {
        return nil, nil;
    }

    limit := assignment.GetSubmissionLimit();
    if (limit == nil) {
        return nil, nil;
    }

    now := time.Now();

    history, err := db.GetSubmissionHistory(assignment, email);
    if (err != nil) {
        return nil, err;
    }

    if (*limit.Max >= 0) {
        if (len(history) >= *limit.Max) {
            return &RejectMaxAttempts{*limit.Max}, nil;
        }
    }

    if (limit.Window != nil) {
        reason, err := checkSubmissionLimitWindow(limit.Window, history, now);
        if (err != nil) {
            return nil, err;
        }

        if (reason != nil) {
            return reason, nil;
        }
    }

    return nil, nil;
}

func checkSubmissionLimitWindow(window *model.SubmittionLimitWindow,
        history []*model.SubmissionHistoryItem, now time.Time) (RejectReason, error) {
    if (len(history) < window.AllowedAttempts) {
        return nil, nil;
    }

    windowStart := now.Add(time.Duration(-window.Duration.TotalNanosecs()));
    earliestTime := time.Time{};

    windowCount := 0;
    for _, item := range history {
        itemTime, err := item.GradingStartTime.Time();
        if (err != nil) {
            return nil, fmt.Errorf("Unable to deserialize submission (%s) time ('%s'): '%w'.", item.ID, item.GradingStartTime, err);
        }

        if (itemTime.After(windowStart)) {
            windowCount++;

            if (earliestTime.IsZero() || earliestTime.After(itemTime)) {
                earliestTime = itemTime;
            }
        }
    }

    if (windowCount >= window.AllowedAttempts) {
        return &RejectWindowMax{window.AllowedAttempts, window.Duration, earliestTime}, nil;
    }

    return nil, nil;
}

func checkLateAcknowledgment(assignment *model.Assignment, lateAcknowledgment bool) RejectReason {
    // if the user acknowledges that they are submitting late, do not reject the submission
    if lateAcknowledgment {
        return nil
    }

    policy := assignment.GetLatePolicy()
    
    if policy.Type == model.EmptyPolicy || policy.Type == model.BaselinePolicy {
        return nil
    }

    dueDate := assignment.DueDate

    if dueDate.IsZero() {
        return nil
    }

    if common.NowTimestamp() >= dueDate {
        return &RejectMissingLateAcknowledgment{}
    }

    return nil
}
