package jplag

import (
	"testing"

	"github.com/edulinq/autograder/internal/analysis/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func TestJPlagComputeFileSimilarityBase(test *testing.T) {
	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.896552,
	}

	engine := GetEngine()

	// Lower the token minimum for testing.
	engineOptsStruct := JPlagEngineOptions{
		MinTokens: 5,
	}

	engineOptions, err := util.ToJSONMap(engineOptsStruct)
	if err != nil {
		test.Errorf("Failed to convert JPlagEngineOption to map[string]any: '%v'.", err)
	}

	core.RunEngineTestComputeFileSimilarityBase(test, engine, false, expected, engineOptions)
}

func TestJPlagComputeFileSimilarityWithIgnoreBase(test *testing.T) {
	expected := &model.FileSimilarity{
		Filename: "submission.py",
		Tool:     NAME,
		Version:  VERSION,
		Score:    0.526316,
	}

	engine := GetEngine()

	// Lower the token minimum for testing.
	engineOptsStruct := JPlagEngineOptions{
		MinTokens: 5,
	}

	engineOptions, err := util.ToJSONMap(engineOptsStruct)
	if err != nil {
		test.Errorf("Failed to convert JPlagEngineOption to map[string]any: '%v'.", err)
	}

	core.RunEngineTestComputeFileSimilarityBase(test, engine, true, expected, engineOptions)
}
