package users

import (
	"slices"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/procedures/users"
)

type EnrollRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	RawCourseUsers []*model.RawCourseUserData `json:"raw-course-users"`

	SkipInserts bool `json:"skip-inserts"`
	SkipUpdates bool `json:"skip-updates"`

	// Send any relevant email (usually about creation or password changing).
	SendEmails bool `json:"send-emails"`

	// Do not actually commit any changes or send any emails regardless of |SendEmails|.
	DryRun bool `json:"dry-run"`
}

type EnrollResponse struct {
	Results []*model.ExternalUserOpResult `json:"results"`
}

func HandleEnroll(request *EnrollRequest) (*EnrollResponse, *core.APIError) {
	rawUsers := model.ToRawUserDatas(request.RawCourseUsers, request.Course)

	options := users.UpsertUsersOptions{
		RawUsers: rawUsers,

		SkipInserts: request.SkipInserts,
		SkipUpdates: request.SkipUpdates,
		SendEmails:  request.SendEmails,
		DryRun:      request.DryRun,

		ContextEmail:      request.ServerUser.Email,
		ContextServerRole: request.ServerUser.Role,
		ContextCourseRole: request.User.Role,
	}

	results := users.UpsertUsers(options)

	var response EnrollResponse
	// Convert UserOpResults to user friendly ExternalUserOpResults.
	for _, result := range results {
		response.Results = append(response.Results, result.ToExternalResult())
	}

	slices.SortFunc(response.Results, model.CompareExternalUserOpResultPointer)

	return &response, nil
}
