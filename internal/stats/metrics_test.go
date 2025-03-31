package stats

import (
	"reflect"
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
			errorSubstring: "No metric was given.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      Unknown_Metric_Attribute_Type,
			},
			errorSubstring: "Metric attribute was not set.",
		},
		{
			metric: &Metric{
				Type: API_Request_Stats_Type,
			},
			errorSubstring: "Metric timestamp was not set.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_Request_Stats_Type,
				Attributes: map[MetricAttribute]any{
					Unknown_Metric_Attribute_Key: "",
				},
			},
			errorSubstring: "Metric attribute key was empty.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_Request_Stats_Type,
				Attributes: map[MetricAttribute]any{
					Course_ID_Key: nil,
				},
			},
			errorSubstring: "Metric attribute value was empty.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_Request_Stats_Type,
				Attributes: map[MetricAttribute]any{
					Course_ID_Key: "",
				},
			},
			errorSubstring: "Metric attribute value was empty.",
		},
		{
			metric: &Metric{
				Timestamp: timestamp.FromMSecs(100),
				Type:      API_Request_Stats_Type,
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

func TestStoreAPIRequestMetric(test *testing.T) {
	metric := Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      API_Request_Stats_Type,
	}

	runStoreStatsTests(test, metric)
}

func TestStoreTaskTimeMetric(test *testing.T) {
	metric := Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      Task_Time_Stats_Type,
	}

	runStoreStatsTests(test, metric)
}

func TestStoreGradingTimeMetric(test *testing.T) {
	metric := Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      Grading_Time_Stats_Type,
	}

	runStoreStatsTests(test, metric)
}

func TestStoreCodeAnalysisTimeMetric(test *testing.T) {
	metric := Metric{
		Timestamp: timestamp.FromMSecs(100),
		Type:      Code_Analysis_Time_Stats_Type,
	}

	runStoreStatsTests(test, metric)
}

func runStoreStatsTests(test *testing.T, metric Metric) {
	clearBackend()
	defer clearBackend()

	if backend != nil {
		test.Fatalf("Stats backend should not be set during testing.")
	}

	typedBackend := makeTestBackend()
	backend = typedBackend

	if len(typedBackend.metric) != 0 {
		test.Fatalf("Found stored stats (%d) before collection.", len(typedBackend.metric))
	}

	AsyncStoreMetric(&metric)

	// Ensure that stats have been collected.
	count := len(typedBackend.metric)
	if count != 1 {
		test.Fatalf("Got an unexpected number of metrics. Expected: 1, Actual: %d.", len(typedBackend.metric))
	}

	// Compare the stored metric with the expected one.
	if !reflect.DeepEqual(util.MustToJSON(metric), util.MustToJSON(typedBackend.metric[0])) {
		test.Fatalf("Stored metric is not as expected. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(metric), util.MustToJSONIndent(typedBackend.metric[0]))
	}
}
