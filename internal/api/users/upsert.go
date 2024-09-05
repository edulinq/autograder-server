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
	Results []*model.UserOpResult `json:"results"`
}

func HandleUpsert(request *UpsertRequest) (*UpsertResponse, *core.APIError) {
	request.ContextEmail = request.UserEmail
	request.ContextServerRole = request.ServerUser.Role

	var response UpsertResponse
	response.Results = users.UpsertUsers(request.UpsertUsersOptions)

	// Clear any potential cleartext passwords before sending the response.
	for i := range response.Results {
		response.Results[i].CleartextPassword = ""
	}

	slices.SortFunc(response.Results, func(a, b *model.UserOpResult) int {
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
