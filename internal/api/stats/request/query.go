package request

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
)

type QueryRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	stats.BaseQuery
}

type QueryResponse struct {
	Records []*stats.RequestMetric `json:"results"`
}

// Query the API request stats for the server.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	records, err := db.GetRequestMetrics(request.BaseQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-301", &request.APIRequestUserContext, "Failed to query request stats.").Err(err)
	}

	response := QueryResponse{
		Records: stats.ApplyBaseQuery(records, request.BaseQuery),
	}

	return &response, nil
}
