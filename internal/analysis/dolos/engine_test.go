package dolos

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestDolosComputeFileSimilarityBase(test *testing.T) {
	// The Dolos container has some strange permission issues when run on Github Actions.
	if config.GITHUB_CI.Get() {
		test.Skip("Skipping on Github Actions.")
	}

	docker.EnsureOrSkipForTest(test)

	engine := GetEngine()
	paths := [2]string{
		filepath.Join(util.RootDirForTesting(), "testdata", "course101", "HW0", "test-submissions", "solution", "submission.py"),
		filepath.Join(util.RootDirForTesting(), "testdata", "course101", "HW0", "test-submissions", "partial", "submission.py"),
	}

	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.666667,
	}

	result, err := engine.ComputeFileSimilarity(paths, "course101")
	if err != nil {
		test.Fatalf("Failed to compute similarity: '%v'.", err)
	}

	// Pull out and normalize the scores, they will be compared separately.
	expectedScore := expected.Score
	actualScore := result.Score

	expected.Score = 0
	result.Score = 0

	if !reflect.DeepEqual(expected, result) {
		test.Fatalf("Result not as expected. Expected: '%s', Actual: '%s'.", util.MustToJSONIndent(expected), util.MustToJSONIndent(result))
	}

	if !util.IsClose(expectedScore, actualScore) {
		test.Fatalf("Score not as expected. Expected: %f, Actual: %f.", expectedScore, actualScore)
	}
}
