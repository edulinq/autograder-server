package core

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

var (
	baseTestRelDir        = filepath.Join("testdata", "files", "sim_engine", "test-submissions")
	solutionRelPath       = filepath.Join(baseTestRelDir, "solution", "submission.py")
	partialRelPath        = filepath.Join(baseTestRelDir, "partial", "submission.py")
	notImplementedRelPath = filepath.Join(baseTestRelDir, "not_implemented", "submission.py")
)

func RunEngineTestComputeFileSimilarityBase(test *testing.T, engine SimilarityEngine, includeTemplate bool, expected *model.FileSimilarity) {
	docker.EnsureOrSkipForTest(test)

	paths := [2]string{
		filepath.Join(util.RootDirForTesting(), solutionRelPath),
		filepath.Join(util.RootDirForTesting(), partialRelPath),
	}

	templatePath := ""
	if includeTemplate {
		templatePath = filepath.Join(util.RootDirForTesting(), notImplementedRelPath)
	}

	result, err := engine.ComputeFileSimilarity(paths, templatePath, context.Background())
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
}
