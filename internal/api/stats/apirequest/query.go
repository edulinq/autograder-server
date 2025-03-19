package apirequest

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
)

type QueryRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	stats.MetricQuery
}

type QueryResponse struct {
	Results []map[string]any `json:"results"`
}

// Query the API request stats for the server.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	records, err := db.GetAPIRequestMetrics(request.MetricQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-301", &request.APIRequestUserContext, "Failed to get API request metrics.").Err(err)
	}

	aggregatedResults, err := stats.QueryAndAggregateMetrics(records, request.MetricQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-302", &request.APIRequestUserContext, err.Error())
	}

	response := QueryResponse{
		Results: aggregatedResults,
	}

	return &response, nil
}
