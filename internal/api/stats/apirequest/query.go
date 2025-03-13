package apirequest

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

type QueryRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	stats.APIRequestMetricQuery

	stats.AggregationQuery
}

// Query the API request stats for the server.
func HandleQuery(request *QueryRequest) (*stats.QueryResponse, *core.APIError) {
	records, err := db.GetAPIRequestMetrics(request.APIRequestMetricQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-301", &request.APIRequestUserContext, "Failed to query API request stats.").Err(err)
	}

	// Convert records to a general format for aggregation.
	records = stats.ApplyBaseQuery(records, request.BaseQuery)
	metrics, err := util.ToJsonMapSlice(records)
	if err != nil {
		return nil, core.NewUserContextInternalError("-302", &request.APIRequestUserContext, "Failed to convert records to a slice of maps.").Err(err)
	}

	queryResponse := stats.QueryResponse{}

	if !request.EnableAggregation {
		queryResponse.Response = metrics
		return &queryResponse, nil
	}

	if request.AggregateField == "" {
		return nil, core.NewBadRequestError("-303", &request.APIRequest, "No aggregate field was supplied.")
	}

	aggregatedResults, err := stats.ApplyAggregation(metrics, stats.APIRequestMetric{}, request.GroupByFields, request.AggregateField)
	if err != nil {
		return nil, core.NewBadRequestError("-304", &request.APIRequest, "Failed to apply aggregation.").Err(err)
	}

	queryResponse.Response = aggregatedResults
	return &queryResponse, nil
}
