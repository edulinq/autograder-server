package dolos

import (
	"runtime"
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/model"
)

func TestDolosComputeFileSimilarityBase(test *testing.T) {
	if runtime.GOARCH != "amd64" {
		test.Skip("Dolos only runs on amd64.")
	}

	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.717949,
	}

	// Empty Engine Option map for testing
	engineOpts := make(model.OptionsMap)

	core.RunEngineTestComputeFileSimilarityBase(test, GetEngine(), false, expected, engineOpts)
}

func TestDolosComputeFileSimilarityWithIgnoreBase(test *testing.T) {
	if runtime.GOARCH != "amd64" {
		test.Skip("Dolos only runs on amd64.")
	}

	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.702703,
	}

	// Empty Engine Option map for testing
	engineOpts := make(model.OptionsMap)

	core.RunEngineTestComputeFileSimilarityBase(test, GetEngine(), true, expected, engineOpts)
}
