package model

import (
    "fmt"
    "strings"
)

const (
    LMS_TYPE_CANVAS = "canvas"
    LMS_TYPE_TEST = "test"
)

type LMSAdapter struct {
    Type string `json:"type"`

    // Connection options.
    LMSCourseID string `json:"course-id"`
    APIToken string `json:"api-token"`
    BaseURL string `json:"base-url"`

    // Behavior options.
    SyncUserAttributes bool `json:"sync-user-attributes"`
    SyncAddUsers bool `json:"sync-add-users"`
    SyncRemoveUsers bool `json:"sync-remove-users"`
}

func (this *LMSAdapter) Validate() error {
    if (this.Type == "") {
        return fmt.Errorf("LMS type cannot be empty.");
    }
    this.Type = strings.ToLower(this.Type);

    return nil;
}
