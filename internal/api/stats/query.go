package stats

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
)

type QueryRequest struct {
	core.APIRequestCourseUserContext
	core.MinServerRoleAdmin

	stats.MetricQuery
}

type QueryResponse struct {
	Records []*stats.BaseMetric `json:"results"`
}

// Query stats for the server.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	records, err := db.GetMetrics(request.MetricQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-301", &request.APIRequestUserContext, "Failed to query stats.").Err(err)
	}

	response := QueryResponse{
		Records: stats.ApplyBaseQuery(records, request.BaseQuery),
	}

	return &response, nil
}
