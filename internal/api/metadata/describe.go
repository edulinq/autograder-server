package metadata

import (
	"github.com/edulinq/autograder/internal/api/core"
)

type DescribeRequest struct {
	core.APIRequest
}

type DescribeResponse struct {
	core.APIDescription
}

// Describe all endpoints on the server.
func HandleDescribe(request *DescribeRequest) (*DescribeResponse, *core.APIError) {
	response := DescribeResponse{core.GetAPIDescription()}

	return &response, nil
}
