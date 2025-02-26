package apirequest

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
)

type QueryRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	stats.APIRequestMetricQuery
}

type QueryResponse struct {
	Records []*stats.APIRequestMetric `json:"results"`
}

// Query the API request stats for the server.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	records, err := db.GetAPIRequestMetrics(request.APIRequestMetricQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-301", &request.APIRequestUserContext, "Failed to query API request stats.").Err(err)
	}

	response := QueryResponse{
		Records: stats.ApplyBaseQuery(records, request.APIRequestMetricQuery.BaseQuery),
	}

	return &response, nil
}
