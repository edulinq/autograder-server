package core

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func RunEngineTestComputeFileSimilarityBase(test *testing.T, engine SimilarityEngine, expected *model.FileSimilarity) {
	docker.EnsureOrSkipForTest(test)

	paths := [2]string{
		filepath.Join(util.RootDirForTesting(), "testdata", "course101", "HW0", "test-submissions", "solution", "submission.py"),
		filepath.Join(util.RootDirForTesting(), "testdata", "course101", "HW0", "test-submissions", "partial", "submission.py"),
	}

	result, runTime, err := engine.ComputeFileSimilarity(paths)
	if err != nil {
		test.Fatalf("Failed to compute similarity: '%v'.", err)
	}

	// Pull out the scores, they will be compared separately.
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

	if runTime <= 0 {
		test.Fatalf("Run time is too small. It should be at least 1 ms, but is %d ms.", runTime)
	}
}
