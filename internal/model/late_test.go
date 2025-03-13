package model

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestLateGradingPolicyValidate(test *testing.T) {
	testCases := []struct {
		input          *LateGradingPolicy
		expected       *LateGradingPolicy
		errorSubstring string
	}{
		{
			&LateGradingPolicy{},
			&LateGradingPolicy{},
			"",
		},
		{
			&LateGradingPolicy{
				Type: EmptyPolicy,
			},
			&LateGradingPolicy{},
			"",
		},
		{
			&LateGradingPolicy{
				Type: BaselinePolicy,
			},
			&LateGradingPolicy{
				Type: BaselinePolicy,
			},
			"",
		},
		{
			&LateGradingPolicy{
				Type:    ConstantPenalty,
				Penalty: 10,
			},
			&LateGradingPolicy{
				Type:    ConstantPenalty,
				Penalty: 10,
			},
			"",
		},
		{
			&LateGradingPolicy{
				Type:    PercentagePenalty,
				Penalty: 0.10,
			},
			&LateGradingPolicy{
				Type:    PercentagePenalty,
				Penalty: 0.10,
			},
			"",
		},
		{
			&LateGradingPolicy{
				Type:          LateDays,
				LateDaysLMSID: "AAA",
				MaxLateDays:   1,
				Penalty:       0.10,
			},
			&LateGradingPolicy{
				Type:          LateDays,
				LateDaysLMSID: "AAA",
				MaxLateDays:   1,
				Penalty:       0.10,
			},
			"",
		},
		{
			&LateGradingPolicy{
				Type:            LateDays,
				LateDaysLMSID:   "AAA",
				MaxLateDays:     1,
				RejectAfterDays: 1,
				Penalty:         0.10,
			},
			&LateGradingPolicy{
				Type:            LateDays,
				LateDaysLMSID:   "AAA",
				MaxLateDays:     1,
				RejectAfterDays: 1,
				Penalty:         0.10,
			},
			"",
		},
		{
			&LateGradingPolicy{
				Type:            LateDays,
				LateDaysLMSID:   "AAA",
				MaxLateDays:     1,
				RejectAfterDays: 10,
				Penalty:         0.10,
			},
			&LateGradingPolicy{
				Type:            LateDays,
				LateDaysLMSID:   "AAA",
				MaxLateDays:     1,
				RejectAfterDays: 10,
				Penalty:         0.10,
			},
			"",
		},

		// Errors

		{
			nil,
			nil,
			"Late policy is nil.",
		},
		{
			&LateGradingPolicy{
				Type: "ZZZ",
			},
			nil,
			"Unknown late policy type",
		},
		{
			&LateGradingPolicy{
				Type:            BaselinePolicy,
				RejectAfterDays: -1,
			},
			nil,
			"should be zero to be ignored or positive to be applied",
		},
		{
			&LateGradingPolicy{
				Type: ConstantPenalty,
			},
			nil,
			"penalty must be larger than zero",
		},
		{
			&LateGradingPolicy{
				Type:    ConstantPenalty,
				Penalty: 0,
			},
			nil,
			"penalty must be larger than zero",
		},
		{
			&LateGradingPolicy{
				Type: PercentagePenalty,
			},
			nil,
			"penalty must be in (0.0, 1.0]",
		},
		{
			&LateGradingPolicy{
				Type:    PercentagePenalty,
				Penalty: 0,
			},
			nil,
			"penalty must be in (0.0, 1.0]",
		},
		{
			&LateGradingPolicy{
				Type:    PercentagePenalty,
				Penalty: 1.1,
			},
			nil,
			"penalty must be in (0.0, 1.0]",
		},
		{
			&LateGradingPolicy{
				Type:          LateDays,
				LateDaysLMSID: "AAA",
				MaxLateDays:   1,
			},
			nil,
			"penalty must be in (0.0, 1.0]",
		},
		{
			&LateGradingPolicy{
				Type:          LateDays,
				LateDaysLMSID: "AAA",
				MaxLateDays:   1,
				Penalty:       0,
			},
			nil,
			"penalty must be in (0.0, 1.0]",
		},
		{
			&LateGradingPolicy{
				Type:          LateDays,
				LateDaysLMSID: "AAA",
				MaxLateDays:   1,
				Penalty:       1.1,
			},
			nil,
			"penalty must be in (0.0, 1.0]",
		},
		{
			&LateGradingPolicy{
				Type:          LateDays,
				LateDaysLMSID: "AAA",
				Penalty:       0.1,
			},
			nil,
			"max late days must be at least 1",
		},
		{
			&LateGradingPolicy{
				Type:          LateDays,
				LateDaysLMSID: "AAA",
				Penalty:       0.1,
				MaxLateDays:   -1,
			},
			nil,
			"max late days must be at least 1",
		},
		{
			&LateGradingPolicy{
				Type:            LateDays,
				LateDaysLMSID:   "AAA",
				Penalty:         0.1,
				MaxLateDays:     10,
				RejectAfterDays: 1,
			},
			nil,
			"max late days must be in [1, <reject days>",
		},
		{
			&LateGradingPolicy{
				Type:        LateDays,
				Penalty:     0.1,
				MaxLateDays: 1,
			},
			nil,
			"LMS ID for late days assignment cannot be empty",
		},
	}

	for i, testCase := range testCases {
		err := testCase.input.Validate()
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to validate '%+v': '%v'.", i, testCase.input, err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error: '%s'.", i, testCase.errorSubstring)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, testCase.input) {
			test.Errorf("Case %d: Result not as expected. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expected), util.MustToJSONIndent(testCase.input))
			continue
		}
	}
}
