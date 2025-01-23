package dolos

import (
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/model"
)

func TestDolosComputeFileSimilarityBase(test *testing.T) {
	expected := &model.FileSimilarity{
		AnalysisFileInfo: model.AnalysisFileInfo{
			Filename: "submission.py",
		},
		Tool:    NAME,
		Version: VERSION,
		Score:   0.666667,
	}

	core.RunEngineTestComputeFileSimilarityBase(test, GetEngine(), expected)
}
