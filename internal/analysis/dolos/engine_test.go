package dolos

import (
	"reflect"
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

	// Empty engine option map for testing.
	engineOptions := make(model.OptionsMap)

	core.RunEngineTestComputeFileSimilarityBase(test, GetEngine(), false, expected, engineOptions)
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

	// Empty engine option map for testing.
	engineOptions := make(model.OptionsMap)

	core.RunEngineTestComputeFileSimilarityBase(test, GetEngine(), true, expected, engineOptions)
}

func TestParseDolosOptions(test *testing.T) {
	testCases := []struct {
		input           map[string]any
		expected        *DolosEngineOptions
		extractionError bool
	}{
		// Empty options.
		{
			input:           nil,
			expected:        nil,
			extractionError: false,
		},
		{
			input:           map[string]any{},
			expected:        GetDefaultDolosOptions(),
			extractionError: false,
		},
		{
			input:           model.OptionsMap{},
			expected:        GetDefaultDolosOptions(),
			extractionError: false,
		},

		// Custom Options.
		{
			input: model.OptionsMap{
				"kgrams-in-window": 15,
				"kgram-length":     23,
			},
			expected: &DolosEngineOptions{
				KGramsInWindow: 15,
				KGramLength:    23,
			},
			extractionError: false,
		},
		{
			input: model.OptionsMap{
				"kgrams-in-window": 16,
			},
			expected: &DolosEngineOptions{
				KGramsInWindow: 16,
				KGramLength:    23,
			},
			extractionError: false,
		},

		// Fallback to Default.
		{
			input: model.OptionsMap{
				"kgrams-in-window": nil,
			},
			expected:        GetDefaultDolosOptions(),
			extractionError: false,
		},

		// Extra Options.
		{
			input: model.OptionsMap{
				"kgrams-in-window": 12,
				"another-option":   "value",
			},
			expected: &DolosEngineOptions{
				KGramsInWindow: 12,
				KGramLength:    23,
			},
			extractionError: false,
		},

		// Errors.
		{
			input: model.OptionsMap{
				"kgrams-in-window": 17.5,
				"kgram-length":     23,
			},
			expected:        nil,
			extractionError: true,
		},
		{
			input: model.OptionsMap{
				"kgrams-in-window": "abc",
			},
			expected:        nil,
			extractionError: true,
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
