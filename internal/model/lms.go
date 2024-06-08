package model

import (
	"fmt"
	"strings"
)

const (
	LMS_TYPE_CANVAS = "canvas"
	LMS_TYPE_TEST   = "test"
)

type LMSAdapter struct {
	Type string `json:"type"`

	// Connection options.
	LMSCourseID string `json:"course-id,omitempty"`
	APIToken    string `json:"api-token,omitempty"`
	BaseURL     string `json:"base-url,omitempty"`

	// Behavior options.

	SyncUserAttributes bool `json:"sync-user-attributes,omitempty"`
	SyncUserAdds       bool `json:"sync-user-adds,omitempty"`
	SyncUserRemoves    bool `json:"sync-user-removes,omitempty"`

	SyncAssignments bool `json:"sync-assignments,omitempty"`
}

type LMSSyncResult struct {
	UserSync       []*UserOpResult       `json:"user-sync"`
	AssignmentSync *AssignmentSyncResult `json:"assignment-sync"`
}

func (this *LMSAdapter) Validate() error {
	if this.Type == "" {
		return fmt.Errorf("LMS type cannot be empty.")
	}
	this.Type = strings.ToLower(this.Type)

	return nil
}

// Returns true if any aspect of users is synced.
func (this *LMSAdapter) SyncUsers() bool {
	return this.SyncUserAttributes || this.SyncUserAdds || this.SyncUserRemoves
}
