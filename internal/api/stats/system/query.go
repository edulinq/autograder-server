package system

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
	Records []*stats.SystemMetrics `json:"results"`
}

// Query the system stats for the server.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	records, err := db.GetSystemStats(request.Query)
	if err != nil {
		return nil, core.NewUserContextInternalError("-300", &request.APIRequestUserContext, "Failed to query system stats.").Err(err)
	}

	response := QueryResponse{
		Records: stats.LimitAndSort(records, request.Query),
	}

	return &response, nil
}
