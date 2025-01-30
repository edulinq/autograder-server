package system

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

type StacksRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin
}

type StacksResponse struct {
	Stacks []*util.CallStack `json:"stacks"`
}

// Get stack traces for all the currently running routines (threads) on the server.
func HandleStacks(request *StacksRequest) (*StacksResponse, *core.APIError) {
	response := StacksResponse{
		Stacks: util.GetAllStackTraces(),
	}

	return &response, nil
}
