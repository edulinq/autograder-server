package user

import (
    "slices"

    "github.com/edulinq/autograder/internal/api/core"
)

type ListRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`
}

type ListResponse struct {
    Users []*core.UserInfo `json:"users"`
}

func HandleList(request *ListRequest) (*ListResponse, *core.APIError) {
    users := make([]*core.UserInfo, 0, len(request.Users));

    for _, user := range request.Users {
        users = append(users, core.NewUserInfo(user));
    }

    slices.SortFunc(users, core.CompareUserInfoPointer);

    return &ListResponse{users}, nil;
}
