package submissions

import (
	"path/filepath"
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

// Use the common main for all tests in this package.
func TestMain(suite *testing.M) {
	core.APITestingMain(suite, GetRoutes())
}

func getTestSubmissionResultPath(shortID string) string {
	return filepath.Join(util.RootDirForTesting(), "testdata", "course101", "submissions", "HW0", "course-student@test.edulinq.org", shortID, "submission-result.json")
}
