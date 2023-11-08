package user

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
)

type RemoveRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    TargetUser core.TargetUser `json:"target-email"`
}

type RemoveResponse struct {
    FoundUser bool `json:"found-user"`
}

func HandleRemove(request *RemoveRequest) (*RemoveResponse, *core.APIError) {
    response := RemoveResponse{};

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;

    if (request.TargetUser.User.Role > request.User.Role) {
        return nil, core.NewBadPermissionsError("-601", &request.APIRequestCourseUserContext, request.TargetUser.User.Role,
                "Cannot remove a user with a higher role.").Add("target-user", request.TargetUser.User.Email);
    }

    _, err := db.RemoveUser(request.Course, request.TargetUser.Email);
    if (err != nil) {
        return nil, core.NewInternalError("-602", &request.APIRequestCourseUserContext,
                "Failed to remove user.").Err(err).Add("email", request.TargetUser.Email);
    }

    return &response, nil;
}
