package grader

import (
    "fmt"
    "time"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
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
    return fmt.Sprintf("Reached the number of max attempts (%d) within submission window (%s). Next allowed submission time: %s.",
            this.Max, this.WindowDuration.ShortString(), nextTime.Format(time.DateTime));
}

func checkForRejection(assignment *model.Assignment, submissionPath string, user string, message string) (RejectReason, error) {
    return checkSubmissionLimit(assignment, user);
}

func checkSubmissionLimit(assignment *model.Assignment, user string) (RejectReason, error) {
    limit := assignment.GetSubmissionLimit();
    if (limit == nil) {
        return nil, nil;
    }

    now := time.Now();

    history, err := db.GetSubmissionHistory(assignment, user);
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
