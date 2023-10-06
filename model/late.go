package model

import (
    "fmt"
    "math"
    "strings"
    "time"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/util"
)

type LateGradingPolicyType string;

const (
    // Apply no late policy at all.
    EmptyPolicy         LateGradingPolicyType = ""
    // Check the baseline (rejection), but nothing else.
    BaselinePolicy            LateGradingPolicyType = "baseline"
    ConstantPenalty     LateGradingPolicyType = "constant-penalty"
    PercentagePenalty   LateGradingPolicyType = "percentage-penalty"
    LateDays            LateGradingPolicyType = "late-days"
)

const (
    LATE_OPTIONS_KEY_PENALTY string = "penalty";
)

type LateGradingPolicy struct {
    Type LateGradingPolicyType `json:"type"`
    Penalty float64 `json:"penalty"`
    RejectAfterDays int `json:"reject-after-days"`

    MaxLateDays int `json"max-late-days"`
    LateDaysCanvasID string `json:"late-days-canvas-id"`
}

func (this *LateGradingPolicy) Validate() error {
    this.Type = LateGradingPolicyType(strings.ToLower(string(this.Type)));

    if (this.RejectAfterDays < 0) {
        return fmt.Errorf("Number of days for rejection is negative (%d), should be zero to be ignored or positive to be applied.", this.RejectAfterDays);
    }

    switch this.Type {
        case EmptyPolicy, BaselinePolicy:
            return nil;
        case ConstantPenalty:
            if (this.Penalty <= 0.0) {
                return fmt.Errorf("Policy '%s': penalty must be larger than zero, found '%s'.", this.Type, util.FloatToStr(this.Penalty));
            }
        case PercentagePenalty:
            if ((this.Penalty <= 0.0) || (this.Penalty > 1.0)) {
                return fmt.Errorf("Policy '%s': penalty must be in (0.0, 1.0], found '%s'.", this.Type, util.FloatToStr(this.Penalty));
            }
        case LateDays:
            if ((this.Penalty <= 0.0) || (this.Penalty > 1.0)) {
                return fmt.Errorf("Policy '%s': penalty must be in (0.0, 1.0], found '%s'.", this.Type, util.FloatToStr(this.Penalty));
            }

            if ((this.MaxLateDays < 1) || (this.MaxLateDays > this.RejectAfterDays)) {
                return fmt.Errorf("Policy '%s': max late days must be in [1, <reject days>(%d)], found '%d'.", this.Type, this.RejectAfterDays, this.MaxLateDays);
            }

            if (this.LateDaysCanvasID == "") {
                return fmt.Errorf("Policy '%s': Canvas ID for late days assignment cannot be empty.", this.Type);
            }
        default:
            return fmt.Errorf("Unknown late policy type: '%s'.", this.Type);
    }

    return nil;
}

// TEST - Need to update assignment comment.

// This assumes that all assignments are in canvas.
func (this *LateGradingPolicy) Apply(
        assignment *Assignment,
        scores map[string]*ScoringInfo,
        dryRun bool) error {
    // Start with each submission getting the raw score.
    for _, score := range scores {
        score.Score = score.RawScore;
    }

    // Empty policy does nothing.
    if (this.Type == EmptyPolicy) {
        return nil;
    }

    canvasAssignment, err := canvas.FetchAssignment(assignment.Course.CanvasInstanceInfo, assignment.CanvasID);
    if (err != nil) {
        return err;
    }

    dueDate, err := time.Parse(time.RFC3339, canvasAssignment.DueDate);
    if (err != nil) {
        return fmt.Errorf("Failed to parse canvas due date '%s': '%w'.", canvasAssignment.DueDate, err);
    }

    this.applyBaselinePolicy(scores, dueDate);

    // Baseline policy is complete.
    if (this.Type == BaselinePolicy) {
        return nil;
    }

    if ((this.Type == ConstantPenalty) || (this.Type == PercentagePenalty)) {
        penalty := this.Penalty;
        if (this.Type == PercentagePenalty) {
            penalty = canvasAssignment.MaxPoints * this.Penalty;
        }

        this.applyConstantPolicy(scores, penalty);
        return nil;
    }

    if (this.Type == LateDays) {
        // TEST
        return fmt.Errorf("Late Days policy not implemented yet.");
    }

    return fmt.Errorf("Unknown late policy type: '%s'.", this.Type);
}

// Apply a common policy.
func (this *LateGradingPolicy) applyBaselinePolicy(scores map[string]*ScoringInfo, dueDate time.Time) {
    for _, score := range scores {
        score.NumDaysLate = lateDays(dueDate, score.SubmissionTime);

        if ((this.RejectAfterDays > 0) && (score.NumDaysLate > this.RejectAfterDays)) {
            score.Reject = true;
            continue;
        }
    }
}

// Apply a constant penalty per late day.
func (this *LateGradingPolicy) applyConstantPolicy(scores map[string]*ScoringInfo, penalty float64) {
    for _, score := range scores {
        if (score.NumDaysLate <= 0) {
            continue;
        }

        score.Score = math.Max(0.0, score.RawScore - (penalty * float64(score.NumDaysLate)));
    }
}

func lateDays(dueDate time.Time, submissionTime time.Time) int {
    if (dueDate.After(submissionTime)) {
        return 0;
    }

    return int(math.Ceil(submissionTime.Sub(dueDate).Hours() / 24.0));
}
