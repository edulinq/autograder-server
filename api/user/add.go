package user

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/usr"
)

type UserGetRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`

    Email string `json:"email"`
}

type UserGetResponse struct {
    FoundUser bool `json:"found"`
    User *userListRow `json:"user"`
}

type userListRow struct {
    Email string `json:"email"`
    Name string `json:"name"`
    Role usr.UserRole `json:"role"`
}

func HandleUserGet(request *UserGetRequest) (*UserGetResponse, *core.APIError) {
    response := UserGetResponse{};

    user := request.Users[request.Email];
    if (user != nil) {
        response.FoundUser = true;
        response.User = &userListRow{
            Email: user.Email,
            Name: user.DisplayName,
            Role: user.Role,
        };
    }

    return &response, nil;
}
