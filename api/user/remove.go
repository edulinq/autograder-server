package user

import (
    "github.com/eriq-augustine/autograder/api/core"
)

type RemoveRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin
    Users core.CourseUsers `json:"-"`

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

    delete(request.Users, request.TargetUser.Email);

    err := request.Course.SaveUsersFile(request.Users);
    if (err != nil) {
        return nil, core.NewInternalError("-602", &request.APIRequestCourseUserContext, "Failed to save users after a remove.");
    }

    return &response, nil;
}
