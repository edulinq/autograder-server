package model

import (
    "fmt"
    "strings"

    "github.com/eriq-augustine/autograder/util"
)

type LateGradingPolicyType string;

const (
    // Apply no late policy at all.
    EmptyPolicy         LateGradingPolicyType = ""
    // Check the baseline (rejection), but nothing else.
    BaselinePolicy      LateGradingPolicyType = "baseline"
    ConstantPenalty     LateGradingPolicyType = "constant-penalty"
    PercentagePenalty   LateGradingPolicyType = "percentage-penalty"
    LateDays            LateGradingPolicyType = "late-days"
)

type LateGradingPolicy struct {
    Type LateGradingPolicyType `json:"type"`
    Penalty float64 `json:"penalty"`
    RejectAfterDays int `json:"reject-after-days"`

    MaxLateDays int `json:"max-late-days"`
    LateDaysLMSID string `json:"late-days-lms-id"`
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

            if (this.LateDaysLMSID == "") {
                return fmt.Errorf("Policy '%s': LMS ID for late days assignment cannot be empty.", this.Type);
            }
        default:
            return fmt.Errorf("Unknown late policy type: '%s'.", this.Type);
    }

    return nil;
}
