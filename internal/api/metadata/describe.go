package metadata

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/log"
)

type DescribeRequest struct {
	core.APIRequest

	ForceCompute bool `json:"force-compute"`
}

type DescribeResponse struct {
	core.APIDescription
}

// Describe all endpoints on the server.
func HandleDescribe(request *DescribeRequest) (*DescribeResponse, *core.APIError) {
	response := DescribeResponse{}

	if !request.ForceCompute {
		apiDescription, err := core.GetAPIDescription()
		if err != nil {
			log.Warn("Unable to get cached API description.", err)
		}

		if apiDescription != nil {
			response.APIDescription = *apiDescription
			return &response, nil
		}
	}

	routes := core.GetAPIRoutes()
	if routes == nil || len(*routes) == 0 {
		return nil, core.NewInternalError("-501", request, "Unable to describe API endpoints when the cached routes are empty.")
	}

	apiDescription, err := core.DescribeRoutes(*routes)
	if err != nil {
		return nil, core.NewInternalError("-502", request, "Failed to describe API endpoints.").Err(err)
	}

	core.SetAPIDescription(apiDescription)

	if apiDescription != nil {
		response.APIDescription = *apiDescription
	}

	return &response, nil
}
