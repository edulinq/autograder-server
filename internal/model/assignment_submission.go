package model

import (
    "fmt"

    "github.com/edulinq/autograder/internal/common"
)

type SubmissionLimitInfo struct {
    Max *int `json:"max-attempts"`
    Window *SubmittionLimitWindow `json:"window,omitempty"`
}

type SubmittionLimitWindow struct {
    AllowedAttempts int `json:"allowed-attempts"`
    Duration common.DurationSpec `json:"duration"`
}

func (this *SubmissionLimitInfo) Validate() error {
    if ((this.Max == nil) || (*this.Max < 0)) {
        value := -1;
        this.Max = &value;
    }

    if (this.Window != nil) {
        err := this.Window.Validate();
        if (err != nil) {
            return err;
        }
    }

    return nil;
}

func (this SubmittionLimitWindow) Validate() error {
    err := this.Duration.Validate();
    if (err != nil) {
        return fmt.Errorf("Submission limit window has invalid duration: '%w'.", err);
    }

    if (this.Duration.TotalNanosecs() == 0) {
        return fmt.Errorf("Submission limit window has duration needs to be non-empty.");
    }

    if (this.AllowedAttempts <= 0) {
        return fmt.Errorf("Submission limit window must have a positive number of allowed attempts, found %d.", this.AllowedAttempts);
    }

    return nil;
}
