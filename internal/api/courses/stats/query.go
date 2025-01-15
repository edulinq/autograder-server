package stats

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
)

type QueryRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	stats.CourseMetricQuery
}

type QueryResponse struct {
	Records []*stats.CourseMetric `json:"results"`
}

// Query metrics for a specific course.
// Only the context course can be queried for, the target-course field will be ignored for this endpoint.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	// The request must be for the given course.
	request.CourseMetricQuery.CourseID = request.Course.ID

	records, err := db.GetCourseMetrics(request.CourseMetricQuery)
	if err != nil {
		return nil, core.NewInternalError("-618", &request.APIRequestCourseUserContext, "Failed to query course stats.").Err(err)
	}

	response := QueryResponse{
		Records: stats.ApplyBaseQuery(records, request.CourseMetricQuery.BaseQuery),
	}

	return &response, nil
}
