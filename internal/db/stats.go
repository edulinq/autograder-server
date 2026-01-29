package db

import (
	"fmt"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

func GetMetrics(query stats.Query) ([]*stats.Metric, error) {
	return GetMetricsFull(query, false)
}

func GetMetricsFull(query stats.Query, useTestingData bool) ([]*stats.Metric, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	if !useTestingData {
		return backend.GetMetrics(query)
	}

	metrics := make([]*stats.Metric, 0, len(TESTING_STATS_METRICS))
	for _, metric := range TESTING_STATS_METRICS {
		if query.Match(metric) {
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

func StoreMetric(record *stats.Metric) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.StoreMetric(record)
}

var TESTING_STATS_METRICS []*stats.Metric = []*stats.Metric{
	&stats.Metric{
		Timestamp: timestamp.Timestamp(1100),
		Type:      stats.MetricTypeAPIRequest,
		Value:     float64(200),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeEndpoint:     fmt.Sprintf("/api/v%02d/courses/assignments/submissions/submit", util.MustGetAPIVersion()),
			stats.MetricAttributeSender:       "127.0.0.1:12345",
			stats.MetricAttributeCourseID:     "course101",
			stats.MetricAttributeAssignmentID: "hw0",
			stats.MetricAttributeUserEmail:    "course-student@test.edulinq.org",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(1200),
		Type:      stats.MetricTypeGradingTime,
		Value:     float64(150),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeCourseID:     "course101",
			stats.MetricAttributeAssignmentID: "hw0",
			stats.MetricAttributeUserEmail:    "course-student@test.edulinq.org",
		},
	},

	&stats.Metric{
		Timestamp: timestamp.Timestamp(1300),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeTaskType: model.TaskTypeCourseBackup,
			stats.MetricAttributeCourseID: "course101",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(1400),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeTaskType: model.TaskTypeCourseEmailLogs,
			stats.MetricAttributeCourseID: "course101",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(1500),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeTaskType: model.TaskTypeCourseReport,
			stats.MetricAttributeCourseID: "course101",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(1600),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeTaskType: model.TaskTypeCourseScoringUpload,
			stats.MetricAttributeCourseID: "course101",
		},
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(1700),
		Type:      stats.MetricTypeTaskTime,
		Value:     float64(100),
		Attributes: map[stats.MetricAttribute]any{
			stats.MetricAttributeTaskType: model.TaskTypeCourseUpdate,
			stats.MetricAttributeCourseID: "course101",
		},
	},

	&stats.Metric{
		Timestamp: timestamp.Timestamp(1800),
		Type:      stats.MetricTypeSystemCPU,
		Value:     float64(5.0),
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(1900),
		Type:      stats.MetricTypeSystemMemory,
		Value:     float64(0.50),
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(2000),
		Type:      stats.MetricTypeSystemNetworkIn,
		Value:     float64(100),
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(2100),
		Type:      stats.MetricTypeSystemNetworkOut,
		Value:     float64(150),
	},

	&stats.Metric{
		Timestamp: timestamp.Timestamp(2200),
		Type:      stats.MetricTypeSystemCPU,
		Value:     float64(10.0),
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(2300),
		Type:      stats.MetricTypeSystemMemory,
		Value:     float64(0.75),
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(2400),
		Type:      stats.MetricTypeSystemNetworkIn,
		Value:     float64(10),
	},
	&stats.Metric{
		Timestamp: timestamp.Timestamp(2500),
		Type:      stats.MetricTypeSystemNetworkOut,
		Value:     float64(200),
	},
}
