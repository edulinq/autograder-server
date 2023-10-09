package web

import (
    "fmt"
    "net/http"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

var MIN_ROLE_USER_ADD model.UserRole = model.Admin;
var MIN_ROLE_USER_GET model.UserRole = model.Grader;
var MIN_ROLE_USER_LIST model.UserRole = model.Grader;
var MIN_ROLE_USER_REMOVE model.UserRole = model.Admin;

// Requests

type UserAddRequest struct {
    BaseAPIRequest
    Email string `json:"email"`
    NewPass string `json:"new-pass"`
    Name string `json:"name"`
    Role string `json:"role"`
    Force bool `json:"force"`
    SendEmail bool `json:"send-email"`
}

type UserAuthRequest struct {
    BaseAPIRequest
    Email string `json:"email"`
    UserPass string `json:"user-pass"`
}

type UserGetRequest struct {
    BaseAPIRequest
    Email string `json:"email"`
}

type UserListRequest struct {
    BaseAPIRequest
}

type UserRemoveRequest struct {
    BaseAPIRequest
    Email string `json:"email"`
}

// Responses

type UserAddResponse struct {
    Success bool `json:"success"`
    UserExists bool `json:"user-exists"`
}

type UserAuthResponse struct {
    Success bool `json:"success"`
    FoundUser bool `json:"found-user"`
}

type UserGetResponse struct {
    User *userListRow `json:"user"`
}

type UserListResponse struct {
    Result []userListRow `json:"result"`
}

type UserRemoveResponse struct {
    FoundUser bool `json:"found-user"`
}

// Misc

type userListRow struct {
    Email string `json:"email"`
    Name string `json:"name"`
    Role model.UserRole `json:"role"`
}

// Constructors

func NewUserAddRequest(request *http.Request) (*UserAddRequest, *APIResponse, error) {
    var apiRequest UserAddRequest;
    err := APIRequestFromPOST(&apiRequest, request);
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
        return nil, NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    if ((user != nil) && (user.Role < MIN_ROLE_USER_ADD)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    if ((apiRequest.NewPass == "") && (!apiRequest.SendEmail)) {
        return nil, NewResponse(http.StatusNotAcceptable, "If a password is not provided, you must send an email."), nil;
    }

    return &apiRequest, nil, nil;
}

func NewUserAuthRequest(request *http.Request) (*UserAuthRequest, *APIResponse, error) {
    var apiRequest UserAuthRequest;
    err := APIRequestFromPOST(&apiRequest, request);
    if (err != nil) {
        return nil, nil, err;
    }

    course := grader.GetCourse(apiRequest.Course);
    if (course == nil) {
        return nil, nil, fmt.Errorf("Unknown course: '%s'.", apiRequest.Course);
    }

    ok, _, err := AuthAPIRequest(&apiRequest.BaseAPIRequest, course);
    if (err != nil) {
        return nil, nil, err;
    } else if (!ok) {
        return nil, NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    return &apiRequest, nil, nil;
}

func NewUserGetRequest(request *http.Request) (*UserGetRequest, *APIResponse, error) {
    var apiRequest UserGetRequest;
    err := APIRequestFromPOST(&apiRequest, request);
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
        return nil, NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    if ((user != nil) && (user.Role < MIN_ROLE_USER_GET)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    return &apiRequest, nil, nil;
}

func NewUserListRequest(request *http.Request) (*UserListRequest, *APIResponse, error) {
    var apiRequest UserListRequest;
    err := APIRequestFromPOST(&apiRequest, request);
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
        return nil, NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    if ((user != nil) && (user.Role < MIN_ROLE_USER_LIST)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    return &apiRequest, nil, nil;
}

func NewUserRemoveRequest(request *http.Request) (*UserRemoveRequest, *APIResponse, error) {
    var apiRequest UserRemoveRequest;
    err := APIRequestFromPOST(&apiRequest, request);
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
        return nil, NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    if ((user != nil) && (user.Role < MIN_ROLE_USER_LIST)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    return &apiRequest, nil, nil;
}

// Close and Clean

func (this *UserAddRequest) Close() error {
    return nil;
}

func (this *UserAddRequest) Clean() error {
    return nil;
}

func (this *UserAuthRequest) Close() error {
    return nil;
}

func (this *UserAuthRequest) Clean() error {
    return nil;
}

func (this *UserGetRequest) Close() error {
    return nil;
}

func (this *UserGetRequest) Clean() error {
    return nil;
}

func (this *UserListRequest) Close() error {
    return nil;
}

func (this *UserListRequest) Clean() error {
    return nil;
}

func (this *UserRemoveRequest) Close() error {
    return nil;
}

func (this *UserRemoveRequest) Clean() error {
    return nil;
}

// Handlers

func handleUserAdd(request *UserAddRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    users, err := course.GetUsers();
    if (err != nil) {
        return 0, nil, err;
    }

    generatedPass := false;
    if (request.NewPass == "") {
        request.NewPass, err = util.RandHex(model.DEFAULT_PASSWORD_LEN);
        if (err != nil) {
            return 0, nil, fmt.Errorf("Failed to generate a default password: '%w'.", err);
        }

        generatedPass = true;
    }

    response := &UserAddResponse{};

    user, userExists, err := model.NewOrMergeUser(users,
        request.Email, request.Name, request.Role, request.NewPass, request.Force);
    response.UserExists = userExists;

    if (err != nil) {
        if (userExists) {
            return 0, response, nil;
        } else {
            return 0, nil, fmt.Errorf("Failed to create a new or merged user: '%w'.", err);
        }
    }

    users[user.Email] = user;

    err = course.SaveUsersFile(users);
    if (err != nil) {
        return 0, nil, err;
    }

    if (request.SendEmail) {
        model.SendUserAddEmail(user, request.NewPass, generatedPass, userExists, false, false);
    }

    return 0, &UserAddResponse{}, nil;
}

func handleUserAuth(request *UserAuthRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    user, err := course.GetUser(request.Email);
    if (err != nil) {
        return 0, nil, err;
    }

    response := &UserAuthResponse{};

    if (user == nil) {
        return 0, response, nil;
    }

    response.FoundUser = true;
    response.Success = user.CheckPassword(request.UserPass);

    return 0, response, nil;
}

func handleUserGet(request *UserGetRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    user, err := course.GetUser(request.Email);
    if (err != nil) {
        return 0, nil, err;
    }

    response := &UserGetResponse{}

    if (user == nil) {
        return 0, response, nil;
    }

    response.User = &userListRow{
        Email: user.Email,
        Name: user.DisplayName,
        Role: user.Role,
    };

    return 0, response, nil;
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
        Result: make([]userListRow, 0, len(users)),
    };

    for _, user := range users {
        response.Result = append(response.Result, userListRow{
            Email: user.Email,
            Name: user.DisplayName,
            Role: user.Role,
        });
    }

    return 0, response, nil;
}

func handleUserRemove(request *UserRemoveRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    users, err := course.GetUsers();
    if (err != nil) {
        return 0, nil, err;
    }

    contextUser := users[request.BaseAPIRequest.User];
    if (contextUser == nil) {
        return 0, nil, fmt.Errorf("Failed to locate context user '%s'.", request.BaseAPIRequest.User);
    }

    response := &UserRemoveResponse{};

    removeUser := users[request.Email];
    if (removeUser == nil) {
        return 0, response, nil;
    }

    if (removeUser.Role > contextUser.Role) {
        return http.StatusForbidden, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    response.FoundUser = true;
    delete(users, removeUser.Email);

    err = course.SaveUsersFile(users);
    if (err != nil) {
        return 0, nil, err;
    }

    return 0, response, nil;
}
