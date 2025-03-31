package db

import (
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestStoreSystemStats(test *testing.T) {
	Clear()
	defer Clear()

	testRecord := stats.SystemMetrics{
		Metric: stats.Metric{
			Timestamp: timestamp.Now(),
		},
		CPUPercent:       1,
		MemPercent:       2,
		NetBytesSent:     3,
		NetBytesReceived: 4,
	}

	query := stats.Query{}

	err := StoreSystemStats(&testRecord)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
	}

	records, err := GetSystemStats(query)
	if err != nil {
		test.Fatalf("Failed to fetch stats: '%v'.", err)
	}

	if len(records) != 1 {
		test.Fatalf("Did not get the correct number of records. Expected: 1, Actual: %d.", len(records))
	}

	if !reflect.DeepEqual(*records[0], testRecord) {
		test.Fatalf("Did not get the expected record back. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(testRecord), util.MustToJSONIndent(*records[0]))
	}
}

func (this *DBTests) DBTestStoreAPIRequestStats(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.API_Request_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.Sender_Key:        "1",
			stats.Endpoint_Key:      "E1",
			stats.User_Email_Key:    "U",
			stats.Assignment_ID_Key: "A",
			stats.Course_ID_Key:     "C",
			stats.Locator_Key:       "11",
		},
	}

	query := stats.Query{
		Type: stats.API_Request_Stats_Type,
	}

	runStoreStatsTests(test, &metric, query)
}

func (this *DBTests) DBTestStoreTaskTimeStats(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.Task_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.User_Email_Key:    "U",
			stats.Assignment_ID_Key: "A",
			stats.Course_ID_Key:     "C",
		},
	}

	query := stats.Query{
		Type: stats.Task_Time_Stats_Type,
	}

	runStoreStatsTests(test, &metric, query)
}

func (this *DBTests) DBTestStoreGradingTimeStats(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.Grading_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.User_Email_Key:    "U",
			stats.Assignment_ID_Key: "A",
			stats.Course_ID_Key:     "C",
		},
	}

	query := stats.Query{
		Type: stats.Grading_Time_Stats_Type,
	}

	runStoreStatsTests(test, &metric, query)
}

func (this *DBTests) DBTestStoreCodeAnalysisTimeStats(test *testing.T) {
	Clear()
	defer Clear()

	metric := stats.Metric{
		Timestamp: timestamp.Now(),
		Type:      stats.Code_Analysis_Time_Stats_Type,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.User_Email_Key:    "U",
			stats.Assignment_ID_Key: "A",
			stats.Course_ID_Key:     "C",
		},
	}

	query := stats.Query{
		Type: stats.Code_Analysis_Time_Stats_Type,
	}

	runStoreStatsTests(test, &metric, query)
}

func (this *DBTests) DBTestStoreMetricFailures(test *testing.T) {
	Clear()
	defer Clear()

	testCases := []struct {
		metric          stats.Metric
		query           stats.Query
		errorSubstring  string
		expectedMetrics []*stats.Metric
	}{
		// Mixed Metric and Query Type.
		{
			metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Type:      stats.Code_Analysis_Time_Stats_Type,
				Value:     float64(100),
				Attributes: map[stats.MetricAttribute]any{
					stats.Course_ID_Key:     "C",
					stats.Assignment_ID_Key: "A",
					stats.User_Email_Key:    "U",
				},
			},
			query: stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			expectedMetrics: []*stats.Metric{},
		},

		// No Metric Type.
		{
			metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Value:     float64(100),
				Attributes: map[stats.MetricAttribute]any{
					stats.Course_ID_Key:     "C",
					stats.Assignment_ID_Key: "A",
					stats.User_Email_Key:    "U",
				},
			},
			query: stats.Query{
				Type: stats.API_Request_Stats_Type,
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			errorSubstring: "No metric type was given",
		},

		// No Query Type.
		{
			metric: stats.Metric{
				Timestamp: timestamp.Now(),
				Value:     float64(100),
				Attributes: map[stats.MetricAttribute]any{
					stats.Course_ID_Key:     "C",
					stats.Assignment_ID_Key: "A",
					stats.User_Email_Key:    "U",
				},
			},
			query: stats.Query{
				Where: map[stats.MetricAttribute]any{
					"course": "C",
				},
			},
			errorSubstring: "No metric type was given",
		},
	}

	for i, testCase := range testCases {
		Clear()

		err := StoreMetric(&testCase.metric)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error outpout. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to store metric '%+v': '%v'.", i, testCase.metric, err)
			}

			continue
		}

		if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error on metric '%v'.", i, util.MustToJSONIndent(testCase.metric))
			continue
		}

		records, err := GetMetrics(testCase.query)
		if err != nil {
			test.Errorf("Case %d: Failed to fetch stats: '%v'.", i, err)
			continue
		}

		if !reflect.DeepEqual(records, testCase.expectedMetrics) {
			test.Errorf("Case %d: Did not get the expected record back. Expected: '%s', Actual: '%s'.",
				i, util.MustToJSONIndent(testCase.expectedMetrics), util.MustToJSONIndent(records))
			continue
		}
	}
}

func runStoreStatsTests(test *testing.T, metric *stats.Metric, query stats.Query) {
	err := StoreMetric(metric)
	if err != nil {
		test.Fatalf("Failed to store stats: '%v'.", err)
	}

	records, err := GetMetrics(query)
	if err != nil {
		test.Fatalf("Failed to fetch stats: '%v'.", err)
	}

	if len(records) != 1 {
		test.Fatalf("Did not get the correct number of records. Expected: 1, Actual: %d.", len(records))
	}

	if !reflect.DeepEqual(records[0], metric) {
		test.Fatalf("Did not get the expected record back. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(metric), util.MustToJSONIndent(*records[0]))
	}
}
