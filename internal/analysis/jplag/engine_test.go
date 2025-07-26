package jplag

import (
	"encoding/json"
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

	engine := GetEngine()
	// Lower the token minimum for testing.
	engineOptsStruct := JPlagEngineOptions{
		MinTokens: 5,
	}

	jsonBytes, err := json.Marshal(engineOptsStruct)
	if err != nil {
		test.Fatalf("Error marshaling JPlagEngineOptions to JSON: '%v'.", err)
	}
	var engineOptions map[string]any
	err = json.Unmarshal(jsonBytes, &engineOptions)
	if err != nil {
		test.Fatalf("Error unmarshaling JSON to map[string]any: '%v'.", err)
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

	jsonBytes, err := json.Marshal(engineOptsStruct)
	if err != nil {
		test.Fatalf("Error marshaling JPlagEngineOptions to JSON: '%v'.", err)
	}
	var engineOptions map[string]any
	err = json.Unmarshal(jsonBytes, &engineOptions)
	if err != nil {
		test.Fatalf("Error unmarshaling JSON to map[string]any: '%v'.", err)
	}

	core.RunEngineTestComputeFileSimilarityBase(test, engine, true, expected, engineOptions)
}
