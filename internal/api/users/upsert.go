package users

import (
	"slices"

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
	Results []*model.ExternalUserOpResult `json:"results"`
}

func HandleUpsert(request *UpsertRequest) (*UpsertResponse, *core.APIError) {
	request.ContextEmail = request.ServerUser.Email
	request.ContextServerRole = request.ServerUser.Role

	results := users.UpsertUsers(request.UpsertUsersOptions)

	var response UpsertResponse
	// Convert UserOpResults to user friendly ExternalUserOpResults.
	for _, result := range results {
		response.Results = append(response.Results, result.ToExternalResult())
	}

	slices.SortFunc(response.Results, model.CompareExternalUserOpResultPointer)

	return &response, nil
}
