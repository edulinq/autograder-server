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
    LMSCourseID string `json:"course-id,omitempty"`
    APIToken string `json:"api-token,omitempty"`
    BaseURL string `json:"base-url,omitempty"`

    // Behavior options.
    SyncUserAttributes bool `json:"sync-user-attributes,omitempty"`
    SyncAddUsers bool `json:"sync-add-users,omitempty"`
    SyncRemoveUsers bool `json:"sync-remove-users,omitempty"`
}

func (this *LMSAdapter) Validate() error {
    if (this.Type == "") {
        return fmt.Errorf("LMS type cannot be empty.");
    }
    this.Type = strings.ToLower(this.Type);

    return nil;
}
