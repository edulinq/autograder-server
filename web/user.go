package web

import (
    "fmt"
    "net/http"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var MIN_ROLE_USER_LIST model.UserRole = model.Grader;

type UserListRequest struct {
    model.BaseAPIRequest
}

type UserListRow struct {
    Email string `json:"email"`
    Name string `json:"name"`
    Role model.UserRole `json:"role"`
}

type UserListResponse struct {
    Result []UserListRow `json:"result"`
}

func (this *UserListRequest) String() string {
    return util.BaseString(this);
}

func NewUserListRequest(request *http.Request) (*UserListRequest, *model.APIResponse, error) {
    var apiRequest UserListRequest;
    err := model.APIRequestFromPOST(&apiRequest, request);
    if (err != nil) {
        return nil, nil, err;
    }

    course := grader.GetCourse(apiRequest.Course);
    if (course == nil) {
        return nil, nil, fmt.Errorf("Unknown course: '%s'.", apiRequest.Course);
    }

    ok, user, err := AuthAPIRequest(&apiRequest.BaseAPIRequest, course);
    if (err != nil) {
        return nil, nil, err;
    } else if (!ok) {
        return nil, model.NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    if ((user != nil) && (user.Role < MIN_ROLE_USER_LIST)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, model.NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    return &apiRequest, nil, nil;
}

func (this *UserListRequest) Close() error {
    return nil;
}

func (this *UserListRequest) Clean() error {
    return nil;
}

func handleUserList(request *UserListRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    users, err := course.GetUsers();
    if (err != nil) {
        return 0, nil, err;
    }

    response := &UserListResponse{
        Result: make([]UserListRow, 0, len(users)),
    };

    for _, user := range users {
        response.Result = append(response.Result, UserListRow{
            Email: user.Email,
            Name: user.DisplayName,
            Role: user.Role,
        });
    }

    return 0, response, nil;
}
