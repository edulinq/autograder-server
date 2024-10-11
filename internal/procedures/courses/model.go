package courses

import (
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

const (
	UNKNOWN_COURSE_ID = "<unknown>"
	DRY_RUN_PREFIX    = "__autograder_dryrun__"
)

type CourseUpsertOptions struct {
	ContextUser *model.ServerUser `json:"-"`

	SkipSourceSync  bool `json:"skip-source-sync"`
	SkipLMSSync     bool `json:"skip-lms-sync"`
	SkipBuildImages bool `json:"skip-build-images"`
	SkipTasks       bool `json:"skip-tasks"`

	DryRun     bool `json:"dry-run"`
	SkipEmails bool `json:"skip-emails"`
}

type CourseUpsertResult struct {
	CourseID string `json:"course-id"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`

	Created bool `json:"created"`
	Updated bool `json:"updated"`

	LMSSyncResult         *model.LMSSyncResult `json:"lms-sync-result"`
	BuiltAssignmentImages []string             `json:"built-assignment-images"`
}

func compareResults(a CourseUpsertResult, b CourseUpsertResult) int {
	return strings.Compare(a.CourseID, b.CourseID)
}
