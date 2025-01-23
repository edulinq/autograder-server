package jplag

import (
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/model"
)

func TestJPlagComputeFileSimilarityBase(test *testing.T) {
	expected := &model.FileSimilarity{
		AnalysisFileInfo: model.AnalysisFileInfo{
			Filename: "submission.py",
		},
		Tool:    NAME,
		Version: VERSION,
		Score:   1.0,
	}

	// Lower the token minimum for testing.
	engine := GetEngine()
	engine.MinTokens = 5

	core.RunEngineTestComputeFileSimilarityBase(test, engine, expected)
}
