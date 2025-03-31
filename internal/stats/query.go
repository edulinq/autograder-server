package stats

import (
	"fmt"
	"slices"

	"github.com/edulinq/autograder/internal/timestamp"
)

// The query for stats.
// Note that the semantics of this struct mean that times before UNIX epoch (negative times)
// must be offset by at least one MS (as a zero value is treated as the end of time).
type Query struct {
	// Limit the number of results.
	// Take this number of results from the top.
	// A non-positive number means no count limit will be applied.
	Limit int `json:"limit"`

	// Only return data from after this time.
	// A value of zero is treated normally here (as UNIX epoch).
	After timestamp.Timestamp `json:"after"`

	// Only return data from before this time.
	// A value of zero is treated as the end of time.
	Before timestamp.Timestamp `json:"before"`

	// Define how the results should be sorted by time.
	// -1 for ascending, 0 for no sorting, 1 for descending.
	Sort int `json:"sort"`

	// Filter results to only include metrics that match Metric attribute field values.
	// Keys are field names (e.g., "course") and values are what to include (e.g., course101).
	// This filter is applied after all other Query conditions are applied.
	Where map[MetricAttribute]any `json:"where,omitempty"`

	// Only return data of this type.
	// This field is required in the query to specify which kind of metric to return.
	Type MetricType `json:"type"`
}

func (this *Query) Validate() error {
	if this == nil {
		return fmt.Errorf("No query was given.")
	}

	if this.Type == Unknown_Metric_Attribute_Type {
		return fmt.Errorf("Query type was not set.")
	}

	if this.Where == nil {
		this.Where = make(map[MetricAttribute]any)
	}

	for field, value := range this.Where {
		if field == Unknown_Metric_Attribute_Key {
			return fmt.Errorf("Query attribute field was empty.")
		}

		if value == nil || value == "" {
			return fmt.Errorf("Query attribute value was empty.")
		}
	}

	return nil
}

func (this *Query) Match(metric *Metric) bool {
	if this == nil {
		return false
	}

	if this.Type != metric.Type {
		return false
	}

	for key, value := range this.Where {
		if value == "" {
			return false
		}

		fieldValue, exists := metric.Attributes[key]
		if !exists || (fieldValue != value) {
			return false
		}
	}

	return true
}

func (this Query) MatchTimeWindow(record TimestampedMetric) bool {
	time := record.GetTimestamp()
	return (this.Before.IsZero() || (time < this.Before)) && (time > this.After)
}

func compareMetric[T TimestampedMetric](order int, a T, b T) int {
	aTime := a.GetTimestamp()
	bTime := b.GetTimestamp()

	if aTime == bTime {
		return 0
	}

	if aTime < bTime {
		return order
	}

	return -order
}

// Apply time-based filtering, sorting, and limiting to a slice of metrics.
// Only metrics matching the time window are returned, with optional sorting and limiting.
func LimitAndSort[T TimestampedMetric](metrics []T, query Query) []T {
	// Ensure the semantics of sort ordering are followed.
	sortOrder := query.Sort
	if sortOrder < 0 {
		sortOrder = -1
	} else if sortOrder > 0 {
		sortOrder = 1
	}

	results := make([]T, 0, len(metrics))

	// First, filter.
	for _, metric := range metrics {
		if query.MatchTimeWindow(metric) {
			results = append(results, metric)
		}
	}

	// Second, sort.
	if query.Sort != 0 {
		slices.SortFunc(results, func(a T, b T) int {
			return compareMetric(sortOrder, a, b)
		})
	}

	// Finally, limit.
	if query.Limit > 0 {
		limit := query.Limit
		if limit > len(results) {
			limit = len(results)
		}

		results = results[0:limit]
	}

	return results
}
