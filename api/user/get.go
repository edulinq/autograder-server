package user

import (
    "github.com/eriq-augustine/autograder/api/core"
)

type UserGetRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`

    TargetEmail string `json:"target-email"`
}

type UserGetResponse struct {
    FoundUser bool `json:"found"`
    User *core.UserInfo `json:"user"`
}

func HandleUserGet(request *UserGetRequest) (*UserGetResponse, *core.APIError) {
    response := UserGetResponse{};

    user := request.Users[request.TargetEmail];
    if (user != nil) {
        response.FoundUser = true;
        response.User = core.NewUserInfo(user);
    }

    return &response, nil;
}
