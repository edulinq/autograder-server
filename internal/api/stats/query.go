package stats

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
)

type QueryRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	stats.Query
}

type QueryResponse struct {
	Records []*stats.Metric `json:"results"`
}

// Query stats for the server.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	err := request.Query.Validate()
	if err != nil {
		return nil, core.NewBadRequestError("-301", &request.APIRequest, "Failed to validate query.").Err(err)
	}

	records, err := db.GetMetrics(request.Query)
	if err != nil {
		return nil, core.NewUserContextInternalError("-302", &request.APIRequestUserContext, "Failed to query stats.").Err(err)
	}

	response := QueryResponse{
		Records: stats.LimitAndSort(records, request.Query),
	}

	return &response, nil
}
