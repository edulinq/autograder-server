package metadata

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type DescribeRequest struct {
	core.APIRequest

	ForceCompute bool `json:"force-compute"`
}

type DescribeResponse struct {
	*core.APIDescription
}

// Describe all endpoints on the server.
func HandleDescribe(request *DescribeRequest) (*DescribeResponse, *core.APIError) {
	apiDescription, err := core.GetAPIDescription(request.ForceCompute)
	if err != nil {
		return nil, core.NewInternalError("-501", request, "Unable to get API description.").Err(err)
	}

	response := DescribeResponse{apiDescription}

	return &response, nil
}
