package user

import (
    "github.com/eriq-augustine/autograder/api/core"
)

type UserGetRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`

    Email string `json:"email"`
}

type UserGetResponse struct {
    FoundUser bool `json:"found"`
    User *UserInfo `json:"user"`
}

func HandleUserGet(request *UserGetRequest) (*UserGetResponse, *core.APIError) {
    response := UserGetResponse{};

    user := request.Users[request.Email];
    if (user != nil) {
        response.FoundUser = true;
        response.User = NewUserInfo(user);
    }

    return &response, nil;
}
