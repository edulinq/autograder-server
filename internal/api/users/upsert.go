package users

import (
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/users"
)

type UpsertRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleAdmin

	users.UpsertUsersOptions
}

type UpsertResponse struct {
	Results []*model.UserOpResponse `json:"results"`
}

func HandleUpsert(request *UpsertRequest) (*UpsertResponse, *core.APIError) {
	request.ContextEmail = request.UserEmail
	request.ContextServerRole = request.ServerUser.Role

	results := users.UpsertUsers(request.UpsertUsersOptions)

	var response UpsertResponse
	// Convert UserOpResults to user friendly UserOpResponses.
	for _, result := range results {
		response.Results = append(response.Results, result.ToResponse())
	}

	slices.SortFunc(response.Results, func(a, b *model.UserOpResponse) int {
		if a == b {
			return 0
		}

		if a == nil {
			return 1
		}

		if b == nil {
			return -1
		}

		return strings.Compare(a.Email, b.Email)
	})

	return &response, nil
}
