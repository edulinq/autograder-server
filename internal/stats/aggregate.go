package stats

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/edulinq/autograder/internal/util"
)

const (
	COUNT    = "count"
	GROUP_BY = "group-by"
	OVERVIEW = "overview"
	STATS    = "stats"
)

type AggregationQuery struct {
	GroupByFields []string `json:"group-by,omitempty"`
	OverviewField string   `json:"overview,omitempty"`
}

func QueryAndAggregateMetrics[T Metric](metrics []T, query MetricQuery) ([]map[string]any, error) {
	metricMapList, err := queryMetrics(metrics, query.BaseQuery)
	if err != nil {
		return nil, fmt.Errorf("Failed to query metrics: '%v'.", err)
	}

	if query.OverviewField == "" && query.GroupByFields != nil {
		return nil, fmt.Errorf("Must include an overview field when grouping by.")
	}

	// Return the queried metrics if the overview and group by fields are empty.
	if query.OverviewField == "" {
		return metricMapList, nil
	}

	metricMapList, err = aggregateMetrics(metrics, query)
	if err != nil {
		return nil, fmt.Errorf("Failed to aggregate metrics: '%v'.", err)
	}

	return metricMapList, nil
}

func queryMetrics[T Metric](metrics []T, query BaseQuery) ([]map[string]any, error) {
	metrics = ApplyBaseQuery(metrics, query)

	return toJSONMetricsMapSlice(metrics)
}

func aggregateMetrics[T Metric](metrics []T, query MetricQuery) ([]map[string]any, error) {
	var metricType T

	err := validateFields(reflect.TypeOf(metricType), query.AggregationQuery)
	if err != nil {
		return nil, err
	}

	metricMapList, err := toJSONMetricsMapSlice(metrics)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert metrics to JSON map slice: '%v'.", err)
	}

	// Group together metrics based on their groupByFields.
	groupedMetricBuckets, err := groupTogetherMetrics(metricMapList, query.AggregationQuery)
	if err != nil {
		return nil, fmt.Errorf("Failed to group metrics: '%v'.", err)
	}

	aggregatedResults := make([]map[string]any, 0, len(groupedMetricBuckets))

	// Aggregate on metrics that are grouped together.
	for _, groupedMetrics := range groupedMetricBuckets {
		result, err := computeGroupAggregation(groupedMetrics, query)
		if err != nil {
			return nil, fmt.Errorf("Failed to compute aggregation: '%v'.", err)
		}

		if result != nil {
			aggregatedResults = append(aggregatedResults, result)
		}
	}

	return aggregatedResults, nil
}

func toJSONMetricsMapSlice[T Metric](metrics []T) ([]map[string]any, error) {
	jsonMetrics, err := util.ToJSON(metrics)
	if err != nil {
		return nil, err
	}

	metricMapList := make([]map[string]any, 0, len(metrics))
	err = util.JSONFromString(jsonMetrics, &metricMapList)
	if err != nil {
		return nil, err
	}

	return metricMapList, nil
}

func validateFields(metricType reflect.Type, query AggregationQuery) error {
	for _, field := range query.GroupByFields {
		if field == query.OverviewField {
			return fmt.Errorf("Group by and overview fields must be different.")
		}
	}

	if metricType.Kind() == reflect.Ptr {
		metricType = metricType.Elem()
	}

	jsonTags := make(map[string]bool)
	extractJSONTags(metricType, jsonTags)

	for _, field := range query.GroupByFields {
		if !jsonTags[field] {
			return fmt.Errorf("Field '%s' is not a valid group-by field.", field)
		}
	}

	if !jsonTags[query.OverviewField] {
		return fmt.Errorf("Field '%s' is not a valid overview field.", query.OverviewField)
	}

	return nil
}

func extractJSONTags(reflectType reflect.Type, jsonTags map[string]bool) {
	for i := range reflectType.NumField() {
		field := reflectType.Field(i)

		// Recursively get all embedded tags.
		if field.Anonymous {
			extractJSONTags(field.Type, jsonTags)
		}

		jsonTag := util.JSONFieldName(field)

		if jsonTag != "" {
			jsonTags[jsonTag] = true
		}
	}
}

// Group together metrics based on their groupByFields.
func groupTogetherMetrics(metrics []map[string]any, query AggregationQuery) (map[string][]map[string]string, error) {
	groupedMetricBuckets := make(map[string][]map[string]string)

	for _, metric := range metrics {
		// Skip grouping this metric if it doesn't contain all groupByFields and the overviewField.
		if !hasAllNeededFields(metric, query) {
			continue
		}

		// Turn all values into strings for aggregation.
		stringMetric := make(map[string]string, len(metric))
		for field, value := range metric {
			stringMetric[field] = fmt.Sprintf("%v", value)
		}

		var groupFieldValues []string
		for _, field := range query.GroupByFields {
			groupFieldValues = append(groupFieldValues, fmt.Sprintf("%v", metric[field]))
		}

		bucketKey := strings.Join(groupFieldValues, "::")
		groupedMetricBuckets[bucketKey] = append(groupedMetricBuckets[bucketKey], stringMetric)
	}

	return groupedMetricBuckets, nil
}

func hasAllNeededFields(metric map[string]any, query AggregationQuery) bool {
	for _, field := range query.GroupByFields {
		_, exists := metric[field]
		if !exists {
			return false
		}
	}

	_, exists := metric[query.OverviewField]
	if !exists {
		return false
	}

	return true
}

// Calculate aggregations for a group of metrics.
func computeGroupAggregation(groupedMetrics []map[string]string, query MetricQuery) (map[string]any, error) {
	if len(groupedMetrics) == 0 {
		return nil, nil
	}

	aggregatedResults, err := extractFields(groupedMetrics[0], query)
	if err != nil {
		return nil, fmt.Errorf("Failed to extract group by fields from metric '%v': '%v'.", groupedMetrics[0], err)
	}

	// Determine if aggregation should be numeric or non-numeric.
	isNumeric := canAggregateAsNumeric(groupedMetrics[0], query.OverviewField)

	if isNumeric {
		return aggregateNumeric(groupedMetrics, aggregatedResults, query.OverviewField)
	} else {
		return aggregateNonNumeric(groupedMetrics, aggregatedResults)
	}
}

func extractFields(metric map[string]string, query MetricQuery) (map[string]any, error) {
	aggregatedResults := make(map[string]any)
	groupByValues := make(map[string]string)

	for _, field := range query.GroupByFields {
		value := metric[field]
		groupByValues[field] = value
	}

	aggregatedResults[GROUP_BY] = groupByValues
	aggregatedResults[OVERVIEW] = query.OverviewField

	return aggregatedResults, nil
}

func canAggregateAsNumeric(metric map[string]string, overviewField string) bool {
	aggregateValue := metric[overviewField]
	_, err := util.StrToFloat(aggregateValue)
	if err == nil {
		return true
	}

	return false
}

func aggregateNumeric(groupedMetrics []map[string]string, aggregateMap map[string]any, overviewField string) (map[string]any, error) {
	var values []float64
	statsMap := make(map[string]any)

	for _, metric := range groupedMetrics {
		aggregateValue := metric[overviewField]

		numericValue, err := util.StrToFloat(aggregateValue)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert string to float: '%v'.", err)
		}

		values = append(values, numericValue)
	}

	stats := util.ComputeAggregates(values)
	statsJSON, err := util.ToJSON(stats)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert stats to JSON: '%v'.", err)
	}

	statsJSONMap, err := util.JSONMapFromString(statsJSON)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert JSON string to JSON map: '%v'.", err)
	}

	for aggregateKey, aggregateValue := range statsJSONMap {
		statsMap[aggregateKey] = aggregateValue
	}

	aggregateMap[STATS] = statsMap

	return aggregateMap, nil
}

func aggregateNonNumeric(groupedMetrics []map[string]string, aggregateMap map[string]any) (map[string]any, error) {
	statsMap := make(map[string]any)

	statsMap[COUNT] = len(groupedMetrics)
	aggregateMap[STATS] = statsMap

	return aggregateMap, nil
}
