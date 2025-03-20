package course

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
)

type QueryRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	stats.MetricQuery
}

type QueryResponse struct {
	Results []map[string]any `json:"results"`
}

// Query metrics for a specific course.
// Only the context course can be queried for.
// Any course specified in the MetricQuery will be ignored for this endpoint.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	if request.Where == nil {
		request.Where = make(map[string]string)
	}

	// The request must be for the context course.
	request.Where[stats.COURSE] = request.Course.ID

	records, err := db.GetCourseMetrics(request.MetricQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-304", &request.APIRequestUserContext, "Failed to get course metrics.").Err(err)
	}

	aggregatedResults, err := stats.QueryAndAggregateMetrics(records, request.MetricQuery)
	if err != nil {
		return nil, core.NewUserContextInternalError("-305", &request.APIRequestUserContext, err.Error())
	}

	response := QueryResponse{
		Results: aggregatedResults,
	}

	return &response, nil
}
