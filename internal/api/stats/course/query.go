package course

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

type QueryRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	stats.CourseMetricQuery
}

// Query metrics for a specific course.
// Only the context course can be queried for, the target-course field will be ignored for this endpoint.
func HandleQuery(request *QueryRequest) (*stats.QueryResponse, *core.APIError) {
	// The request must be for the given course.
	request.IncludeCourseMetricField.CourseID = request.Course.ID

	records, err := db.GetCourseMetrics(request.CourseMetricQuery)
	if err != nil {
		return nil, core.NewInternalError("-305", &request.APIRequestCourseUserContext, "Failed to query course stats.").Err(err)
	}

	// Convert records to a general format for aggregation.
	records = stats.ApplyBaseQuery(records, request.CourseMetricQuery.BaseQuery)
	metrics, err := util.ToJsonMapSlice(records)
	if err != nil {
		return nil, core.NewUserContextInternalError("-306", &request.APIRequestUserContext, "Failed to convert records to a slice of maps.").Err(err)
	}

	queryResponse := stats.QueryResponse{}

	if !request.EnableAggregation {
		queryResponse.Response = metrics
		return &queryResponse, nil
	}

	if request.AggregateField == "" {
		return nil, core.NewBadRequestError("-307", &request.APIRequest, "No aggregate field was supplied.")
	}

	aggregatedResults, err := stats.ApplyAggregation(metrics, stats.APIRequestMetric{}, request.GroupByFields, request.AggregateField)
	if err != nil {
		return nil, core.NewBadRequestError("-308", &request.APIRequest, "Failed to apply aggregation.").Err(err)
	}

	queryResponse.Response = aggregatedResults
	return &queryResponse, nil
}
