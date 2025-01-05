package courses

import (
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

const (
	UNKNOWN_COURSE_ID = "<unknown>"
	DRY_RUN_PREFIX    = "__autograder_dryrun__"
)

type CourseUpsertPublicOptions struct {
	SkipSourceSync  bool `json:"skip-source-sync" help:"Skip syncing the course's source." default:"false"`
	SkipLMSSync     bool `json:"skip-lms-sync" help:"Skip syncing with the course's LMS." default:"false"`
	SkipBuildImages bool `json:"skip-build-images" help:"Skip building the course's assignment images." default:"false"`
	SkipTasks       bool `json:"skip-tasks" help:"Skip starting up the course's tasks." default:"false"`

	DryRun     bool `json:"dry-run" help:"Do not actually do the operation, just state what you would do." default:"false"`
	SkipEmails bool `json:"skip-emails" help:"Skip sending out emails (always true if a dry run)." default:"false"`
}

type CourseUpsertOptions struct {
	ContextUser *model.ServerUser `json:"-"`

	CourseUpsertPublicOptions
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
