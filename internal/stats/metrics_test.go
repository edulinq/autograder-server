package stats

import (
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func TestMetricValidationBase(test *testing.T) {
	defer clearBackend()

	testCases := []struct {
		metric         *Metric
		errorSubstring string
	}{
		{
			metric:         nil,
			errorSubstring: "No metric was given.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      UnknownMetricAttributeType,
			},
			errorSubstring: "Type was not set.",
		},
		{
			metric: &Metric{
				Type: APIRequestStatsType,
			},
			errorSubstring: "Metric timestamp was not set.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      APIRequestStatsType,
				Attributes: map[MetricAttribute]any{
					UnknownMetricAttributeKey: "",
				},
			},
			errorSubstring: "Attribute key was empty.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      APIRequestStatsType,
				Attributes: map[MetricAttribute]any{
					CourseIDKey: nil,
				},
			},
			errorSubstring: "Attribute value was empty for key 'course'.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      APIRequestStatsType,
				Attributes: map[MetricAttribute]any{
					CourseIDKey: "",
				},
			},
			errorSubstring: "Attribute value was empty for key 'course'.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      APIRequestStatsType,
			},
		},
	}

	for i, testCase := range testCases {
		err := testCase.metric.Validate()
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to validate metric '%+v': '%v'.", i, util.MustToJSONIndent(testCase.metric), err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error on metric '%+v'.", i, util.MustToJSONIndent(testCase.metric))
			continue
		}
	}
}
