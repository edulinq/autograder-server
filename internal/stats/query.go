package stats

import (
	"slices"

	"github.com/edulinq/autograder/internal/timestamp"
)

type Query interface {
	// Does this metric match the filtering conditions of this query.
	Match(record Metric) bool
}

// The base for a stats query.
// Note that the semantics of this struct mean that times before UNIX epoch (negative times)
// must be offset by at least one MS (as a zero value is treated as the end of time).
type BaseQuery struct {
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
}

func (this BaseQuery) Match(record Metric) bool {
	time := record.GetTimestamp()
	return (this.Before.IsZero() || (time < this.Before)) && (time > this.After)
}

func compareMetric[T Metric](order int, a T, b T) int {
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

// Apply the base query to filter the given metrics.
// Return a new list with only the query results.
// The given metrics must already be sorted.
func ApplyBaseQuery[T Metric](metrics []T, baseQuery BaseQuery) []T {
	// Ensure the semantics of sort ordering are followed.
	sortOrder := baseQuery.Sort
	if sortOrder < 0 {
		sortOrder = -1
	} else if sortOrder > 0 {
		sortOrder = 1
	}

	results := make([]T, 0, len(metrics))

	// First, filter.
	for _, metric := range metrics {
		if baseQuery.Match(metric) {
			results = append(results, metric)
		}
	}

	// Second, sort.
	if baseQuery.Sort != 0 {
		slices.SortFunc(results, func(a T, b T) int {
			return compareMetric(sortOrder, a, b)
		})
	}

	// Finally, limit.
	if baseQuery.Limit > 0 {
		limit := baseQuery.Limit
		if limit > len(results) {
			limit = len(results)
		}

		results = results[0:limit]
	}

	return results
}
