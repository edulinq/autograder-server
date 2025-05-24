package stats

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
)

type QueryRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	stats.Query
}

type QueryResponse struct {
	Records []*stats.Metric `json:"results"`
}

// Query metrics for a specific course.
// Only the context course can be queried for, the target-course field will be ignored for this endpoint.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	err := request.Query.Validate()
	if err != nil {
		return nil, core.NewBadRequestError("-618", request, "Failed to validate query.").Err(err)
	}

	// The request must be for the given course.
	request.Query.Where[stats.MetricAttributeCourseID] = request.Course.ID

	records, err := db.GetMetrics(request.Query)
	if err != nil {
		return nil, core.NewInternalError("-630", request, "Failed to query course stats.").Err(err)
	}

	response := QueryResponse{
		Records: stats.LimitAndSort(records, request.Query),
	}

	return &response, nil
}
