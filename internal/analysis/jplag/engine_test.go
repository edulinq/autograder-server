package jplag

import (
	"reflect"
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

func TestParseJplagOptions(test *testing.T) {
	testCases := []struct {
		input           map[string]any
		expected        *JPlagEngineOptions
		extractionError bool
	}{
		// Base Case
		{
			input:           model.OptionsMap{"min-tokens": 100},
			expected:        &JPlagEngineOptions{MinTokens: 100},
			extractionError: false,
		},

		// Empty options
		{
			input:           map[string]any{},
			expected:        GetDefaultJPlagOptions(),
			extractionError: false,
		},
		{
			input:           model.OptionsMap{},
			expected:        GetDefaultJPlagOptions(),
			extractionError: false,
		},

		// Errors
		{
			input:           model.OptionsMap{"min-tokens": 75.5},
			expected:        nil,
			extractionError: true,
		},
		{
			input:           model.OptionsMap{"min-tokens": "abc"},
			expected:        nil,
			extractionError: true,
		},

		// Fallback to Default
		{
			input:           model.OptionsMap{"min-tokens": nil},
			expected:        GetDefaultJPlagOptions(),
			extractionError: false,
		},

		// Extra Options
		{
			input:           model.OptionsMap{"min-tokens": 200, "another-option": "value"},
			expected:        &JPlagEngineOptions{MinTokens: 200},
			extractionError: false,
		},
	}

	for i, testCase := range testCases {
		effectiveOptions, err := parseEngineOptions(testCase.input)
		if err != nil {
			if !testCase.extractionError {
				test.Errorf("Case %d: Got an unexpected error: '%v'.", i, err)
				continue
			}
		} else {
			if testCase.extractionError {
				test.Errorf("Case %d: Did not get an expected error.", i)
				continue
			}
		}

		if !reflect.DeepEqual(effectiveOptions, testCase.expected) {
			test.Errorf("Case %d: Unexpected result. Expected = '%v', Actual = '%v'.", i, testCase.expected, effectiveOptions)
			continue
		}
	}
}
