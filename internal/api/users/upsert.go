package users

import (
    "github.com/edulinq/autograder/internal/api/core"
    "github.com/edulinq/autograder/internal/model"
    "github.com/edulinq/autograder/internal/procedures/users"
)

type UpsertRequest struct {
    core.APIRequestUserContext
    core.MinServerRoleAdmin

    RawUsers []model.RawUserData `json:"raw-users"`

    CourseID string `json:"course-id"`

    SkipInserts bool `json:"skip-inserts"`
    SkipUpdates bool `json:"skip-updates"`

    SendEmails bool `json:send-emails`

    DryRun bool `json:dry-run`
}

type UpsertResponse struct {
    Result []*model.UserOpResult `json:"result"`
}

func HandleUpsert(request *UpsertRequest) (*UpsertResponse, *core.APIError) {
    rawUsers := make([]*model.RawUserData, 0, len(request.RawUsers))

    for _, rawUser := range request.RawUsers {
        rawUsers = append(rawUsers, &rawUser)
    }

    options := users.UpsertUsersOptions{
        RawUsers: rawUsers,
        SkipInserts: request.SkipInserts,
        SkipUpdates: request.SkipUpdates,
        SendEmails: request.SendEmails,
        DryRun: request.DryRun,
        ContextEmail: request.UserEmail,
        ContextServerRole: request.ServerUser.Role,
        ContextCourseRole: request.ServerUser.GetCourseRole(request.CourseID),
    }

    var response UpsertResponse
    response.Result = users.UpsertUsers(options)

    return response
}
