package api

import (
    "fmt"

    "github.com/eriq-augustine/autograder/usr"
)

type UserGetRequest struct {
    APIRequestCourseUserContext
    MinRoleGrader
    Users CourseUsers `json:"-"`

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

func handleUserGet(request *UserGetRequest) (*UserGetResponse, *APIError) {
    // TEST
    // return nil, nil;

    // TEST
    fmt.Println("TEST - HANDLE");
    fmt.Println(request.course);
    fmt.Println(request.Users);

    return &UserGetResponse{
        FoundUser: true,
        User: &userListRow{"X", "Y", usr.Owner},
    }, nil;
}
