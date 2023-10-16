package user

import (
    "github.com/eriq-augustine/autograder/api/core"
)

type UserListRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`
}

type UserListResponse struct {
    Users []*core.UserInfo `json:"users"`
}

func HandleUserList(request *UserListRequest) (*UserListResponse, *core.APIError) {
    users := make([]*core.UserInfo, 0, len(request.Users));

    for _, user := range request.Users {
        users = append(users, core.NewUserInfo(user));
    }

    return &UserListResponse{users}, nil;
}
