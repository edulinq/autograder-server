package apirequest

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/stats"
	"github.com/edulinq/autograder/internal/util"
)

type AggregateRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	stats.APIRequestMetricAggregate
}

type AggregateResponse struct {
	Records *map[string]util.AggregateValues `json:"results"`
}

// Aggregate and filter API request metrics.
func HandleAggregate(request *AggregateRequest) (*AggregateResponse, *core.APIError) {
	if request.APIRequestMetricAggregate.GroupBy == "" {
		return nil, core.NewBadRequestError("-302", &core.APIRequest{}, "No group by specified.")
	}

	records, err := db.GetFilteredAPIRequestMetrics(request.APIRequestMetricAggregate)
	if err != nil {
		return nil, core.NewUserContextInternalError("-303", &request.APIRequestUserContext, "Failed to aggregate API request stats.").Err(err)
	}

	records = stats.ApplyBaseQuery(records, request.APIRequestMetricAggregate.BaseQuery)

	response := AggregateResponse{
		Records: stats.ApplyAggregate(records, request.APIRequestMetricAggregate.GroupBy),
	}

	return &response, nil
}
