package model

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

type LateGradingPolicyType string

const (
	// Apply no late policy at all.
	EmptyPolicy LateGradingPolicyType = ""
	// Check the baseline (rejection), but nothing else.
	BaselinePolicy    LateGradingPolicyType = "baseline"
	ConstantPenalty   LateGradingPolicyType = "constant-penalty"
	PercentagePenalty LateGradingPolicyType = "percentage-penalty"
	LateDays          LateGradingPolicyType = "late-days"
)

type LateGradingPolicy struct {
	Type            LateGradingPolicyType `json:"type"`
	Penalty         float64               `json:"penalty,omitempty"`
	RejectAfterDays int                   `json:"reject-after-days,omitempty"`
	GraceMinutes    int                   `json:"grace-mins,omitempty"`

	MaxLateDays     int    `json:"max-late-days,omitempty"`
	LateDaysLMSID   string `json:"late-days-lms-id,omitempty"`
	LateDaysLMSName string `json:"late-days-lms-name,omitempty"`
}

func (this *LateGradingPolicy) Validate() error {
	if this == nil {
		return fmt.Errorf("Late policy is nil.")
	}

	this.Type = LateGradingPolicyType(strings.ToLower(string(this.Type)))

	if this.RejectAfterDays < 0 {
		return fmt.Errorf("Number of days for rejection is negative (%d), should be zero to be ignored or positive to be applied.", this.RejectAfterDays)
	}

	if this.GraceMinutes < 0 {
		return fmt.Errorf("Grace time in minutes is negative (%d), should be zero to be ignored or positive to be applied.", this.GraceMinutes)
	}

	switch this.Type {
	case EmptyPolicy, BaselinePolicy:
		return nil
	case ConstantPenalty:
		if this.Penalty <= 0.0 {
			return fmt.Errorf("Policy '%s': penalty must be larger than zero, found '%s'.", this.Type, util.FloatToStr(this.Penalty))
		}
	case PercentagePenalty:
		if (this.Penalty <= 0.0) || (this.Penalty > 1.0) {
			return fmt.Errorf("Policy '%s': penalty must be in (0.0, 1.0], found '%s'.", this.Type, util.FloatToStr(this.Penalty))
		}
	case LateDays:
		if (this.Penalty <= 0.0) || (this.Penalty > 1.0) {
			return fmt.Errorf("Policy '%s': penalty must be in (0.0, 1.0], found '%s'.", this.Type, util.FloatToStr(this.Penalty))
		}

		if this.MaxLateDays < 1 {
			return fmt.Errorf("Policy '%s': max late days must be at least 1, found '%d'.", this.Type, this.MaxLateDays)
		}

		if (this.RejectAfterDays > 0) && (this.MaxLateDays > this.RejectAfterDays) {
			return fmt.Errorf("Policy '%s': max late days must be in [1, <reject days>(%d)], found '%d'.", this.Type, this.RejectAfterDays, this.MaxLateDays)
		}

		if (this.LateDaysLMSID == "") && (this.LateDaysLMSName == "") {
			return fmt.Errorf("Policy '%s': Both LMS ID and name for late days assignment cannot be empty.", this.Type)
		}
	default:
		return fmt.Errorf("Unknown late policy type: '%s'.", this.Type)
	}

	return nil
}
