package system

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

type QueryRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	stats.BaseQuery

	stats.AggregationQuery
}

type QueryResponse struct {
	Records []*stats.SystemMetrics `json:"results"`
}

// Query the system stats for the server.
func HandleQuery(request *QueryRequest) (*stats.QueryResponse, *core.APIError) {
	records, err := db.GetSystemStats(request.BaseQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-304", &request.APIRequestUserContext, "Failed to query system stats.").Err(err)
	}

	records = stats.ApplyBaseQuery(records, request.BaseQuery)

	metrics, err := util.ToJsonMapSlice(records)
	if err != nil {
		return nil, core.NewUserContextInternalError("-305", &request.APIRequestUserContext, "Failed to convert records to a slice of maps.").Err(err)
	}

	queryResponse := stats.QueryResponse{}

	if !request.EnableAggregation {
		queryResponse.Response = metrics
		return &queryResponse, nil
	}

	if request.AggregateField == "" {
		return nil, core.NewBadRequestError("-306", &request.APIRequest, "No aggregate field supplied.")
	}

	aggregateResults, err := stats.ApplyAggregation(metrics, stats.APIRequestMetric{}, request.GroupByFields, request.AggregateField)
	if err != nil {
		return nil, core.NewBadRequestError("-307", &request.APIRequest, "Failed to apply aggregation.").Err(err)
	}

	queryResponse.Response = aggregateResults

	return &queryResponse, nil
}
