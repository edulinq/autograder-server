package stats

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

type AggregationQuery struct {
	EnableAggregation bool     `json:"enable-aggregation"`
	GroupByFields     []string `json:"group-by"`
	AggregateField    string   `json:"aggregate"`
}

type QueryResponse struct {
	Response []map[string]any `json:"response"`
}

const (
	COUNT_KEY  = "count"
	MAX_KEY    = "max"
	MEAN_KEY   = "mean"
	MEDIAN_KEY = "median"
	MIN_KEY    = "min"
)

// Aggregate metrics based on groupByFields and the aggregateField.
func ApplyAggregation(metrics []map[string]any, metricType any, groupByFields []string, aggregateField string) ([]map[string]any, error) {
	err := validate(metricType, groupByFields, aggregateField)
	if err != nil {
		return nil, fmt.Errorf("Failed to validate data for aggregation: '%v'.", err)
	}

	// Group together metrics based on their groupByFields.
	groupedMetricBuckets, err := groupTogetherMetrics(metrics, groupByFields, aggregateField)
	if err != nil {
		return nil, fmt.Errorf("Failed to group metrics: '%v'.", err)
	}

	aggregatedResults := make([]map[string]any, 0, len(groupedMetricBuckets))

	// Aggregate on metrics that are grouped together.
	for _, groupedMetrics := range groupedMetricBuckets {
		result, err := computeGroupAggregation(groupedMetrics, aggregateField, groupByFields)
		if err != nil {
			return nil, fmt.Errorf("Failed to compute aggregations: '%v'.", err)
		}

		if result != nil {
			aggregatedResults = append(aggregatedResults, result)
		}
	}

	return aggregatedResults, nil
}

func extractJSONTags(reflectType reflect.Type, jsonTags map[string]bool) {
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)

		// Recursively get all embedded structs.
		if field.Anonymous {
			extractJSONTags(field.Type, jsonTags)
		}

		jsonTag := util.JSONFieldName(field)

		if jsonTag != "" {
			jsonTags[jsonTag] = true
		}
	}
}

// Ensure the groupByFields and aggregateFields exist in the metricType.
func validate(metricType any, groupByFields []string, aggregateField string) error {
	reflectType := reflect.ValueOf(metricType).Type()

	jsonTags := make(map[string]bool)
	extractJSONTags(reflectType, jsonTags)

	for _, field := range groupByFields {
		if !jsonTags[field] {
			return fmt.Errorf("Failed to validate aggregation query. '%s' is not a valid group-by field.", field)
		}
	}

	if !jsonTags[aggregateField] {
		return fmt.Errorf("Failed to validate aggregation query. '%s' is not a valid aggregate field.", aggregateField)
	}

	return nil
}

// Group together metrics based on their groupByFields.
func groupTogetherMetrics(metrics []map[string]any, groupByFields []string, aggregateField string) (map[string][]map[string]any, error) {
	groupedMetricBuckets := make(map[string][]map[string]any)

	for _, metric := range metrics {
		// Skip grouping this metric if it doesn't contain all groupByFields and the aggregateField.
		if !hasAllNeededFields(metric, groupByFields, aggregateField) {
			continue
		}

		var groupFieldValues []string
		for _, field := range groupByFields {
			groupFieldValues = append(groupFieldValues, fmt.Sprintf("%v", metric[field]))
		}

		bucketKey := strings.Join(groupFieldValues, "|")
		groupedMetricBuckets[bucketKey] = append(groupedMetricBuckets[bucketKey], metric)
	}

	return groupedMetricBuckets, nil
}

func hasAllNeededFields(metric map[string]any, groupByFields []string, aggregateField string) bool {
	for _, field := range groupByFields {
		_, exists := metric[field]
		if !exists {
			return false
		}
	}

	_, exists := metric[aggregateField]
	if !exists {
		return false
	}

	return true
}

// Calculate aggregations for a group of metrics.
func computeGroupAggregation(groupedMetrics []map[string]any, aggregateField string, groupByFields []string) (map[string]any, error) {
	if len(groupedMetrics) == 0 {
		return nil, nil
	}

	aggregatedResults := make(map[string]any)

	// Use the first metric in the group to extract the group-by field values.
	firstMetric := groupedMetrics[0]
	for _, field := range groupByFields {
		value, exists := firstMetric[field]
		if !exists {
			return nil, fmt.Errorf("Failed to get the group-by field '%s' for metric '%v'.", field, firstMetric)
		}

		aggregatedResults[field] = value
	}

	var values []float64
	var totalCount int = 0

	for _, metric := range groupedMetrics {
		aggregateValue, exists := metric[aggregateField]
		if !exists {
			// Skip aggregating this metric if it doesn't have the aggregation field.
			continue
		}

		totalCount++

		aggregateValueString := fmt.Sprintf("%v", aggregateValue)
		numericValue, err := util.StrToFloat(aggregateValueString)
		// Only append to values if the aggregateValue was a number.
		if err == nil {
			values = append(values, numericValue)
		}
	}

	if totalCount == 0 {
		return nil, nil
	}

	aggregatedResults[COUNT_KEY] = totalCount

	// Only add additional stats if aggregation was done on numeric values.
	if len(values) > 0 {
		stats := util.ComputeAggregates(values)

		aggregatedResults[MEAN_KEY] = stats.Mean
		aggregatedResults[MEDIAN_KEY] = stats.Median
		aggregatedResults[MIN_KEY] = stats.Min
		aggregatedResults[MAX_KEY] = stats.Max
	}

	return aggregatedResults, nil
}

func SortFunc(a, b any) int {
	jsonA, _ := json.Marshal(a)
	jsonB, _ := json.Marshal(b)

	if string(jsonA) < string(jsonB) {
		return -1
	} else if string(jsonA) > string(jsonB) {
		return 1
	}
	return 0
}
