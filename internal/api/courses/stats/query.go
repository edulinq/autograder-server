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
	if request.Where == nil {
		request.Where = make(map[stats.MetricAttribute]any)
	}

	// The request must be for the given course.
	request.Where[stats.COURSE_ID_KEY] = request.Course.ID

	records, err := db.GetMetrics(request.Query)
	if err != nil {
		return nil, core.NewInternalError("-618", &request.APIRequestCourseUserContext, "Failed to query course stats.").Err(err)
	}

	response := QueryResponse{
		Records: stats.LimitAndSort(records, request.Query),
	}

	return &response, nil
}
