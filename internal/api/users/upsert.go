package users

import (
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
	Results []*model.UserOpResult `json:"results"`
}

func HandleUpsert(request *UpsertRequest) (*UpsertResponse, *core.APIError) {
	request.ContextEmail = request.UserEmail
	request.ContextServerRole = request.ServerUser.Role

	var response UpsertResponse
	response.Results = users.UpsertUsers(request.UpsertUsersOptions)

	return &response, nil
}
