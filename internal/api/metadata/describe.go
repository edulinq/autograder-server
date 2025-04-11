package metadata

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type DescribeRequest struct {
	core.APIRequest

	ForceUpdate bool `json:"force-update"`
}

type DescribeResponse struct {
	core.APIDescription
}

// Describe all endpoints on the server.
func HandleDescribe(request *DescribeRequest) (*DescribeResponse, *core.APIError) {
	apiDescription := core.GetAPIDescription()

	var err error
	if apiDescription == nil || request.ForceUpdate {
		routes := core.GetAPIRoutes()
		if len(routes) == 0 {
			return nil, core.NewInternalError("-501", request, "Unable to describe API endpoints because the cached routes are empty.")
		}

		apiDescription, err = core.Describe(routes)
		if err != nil {
			return nil, core.NewInternalError("-502", request, "Failed to describe API endpoints.").Err(err)
		}
	}

	core.SetAPIDescription(apiDescription)

	response := DescribeResponse{*apiDescription}

	return &response, nil
}
