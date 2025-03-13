package jplag

import (
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/model"
)

func TestJPlagComputeFileSimilarityBase(test *testing.T) {
	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.896552,
	}

	// Lower the token minimum for testing.
	engine := GetEngine()
	engine.MinTokens = 5

	core.RunEngineTestComputeFileSimilarityBase(test, engine, false, expected)
}

func TestJPlagComputeFileSimilarityWithIgnoreBase(test *testing.T) {
	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.526316,
	}

	// Lower the token minimum for testing.
	engine := GetEngine()
	engine.MinTokens = 5

	core.RunEngineTestComputeFileSimilarityBase(test, engine, true, expected)
}
