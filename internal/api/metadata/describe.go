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

func HandleDescribe(request *DescribeRequest) (*DescribeResponse, *core.APIError) {
	response := DescribeResponse{*core.Describe(core.GetServerRoutes())}

	return &response, nil
}
