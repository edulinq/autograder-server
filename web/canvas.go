package web

import (
    "fmt"
    "net/http"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
)

var MIN_ROLE_CANVAS_SYNC_USERS model.UserRole = model.Admin;

// Requests

type CanvasSyncUsersRequest struct {
    model.BaseAPIRequest
}

// Responses

type CanvasSyncUsersResponse struct {
    Success bool `json:"success"`
    Count int `json:"count"`
}

// Constructors

func NewCanvasSyncUsersRequest(request *http.Request) (*CanvasSyncUsersRequest, *model.APIResponse, error) {
    var apiRequest CanvasSyncUsersRequest;
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

    if ((user != nil) && (user.Role < MIN_ROLE_CANVAS_SYNC_USERS)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, model.NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    if (course.CanvasInstanceInfo == nil) {
        return nil, nil, fmt.Errorf("Course '%s' does not have any canvas instance information.", apiRequest.Course);
    }

    return &apiRequest, nil, nil;
}

// Close and Clean

func (this *CanvasSyncUsersRequest) Close() error {
    return nil;
}

func (this *CanvasSyncUsersRequest) Clean() error {
    return nil;
}

// Handlers

func handleCanvasSyncUsers(request *CanvasSyncUsersRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    count, err := course.SyncCanvasUsers();
    if (err != nil) {
        return 0, nil, err;
    }

    response := &CanvasSyncUsersResponse{
        Success: true,
        Count: count,
    };

    return 0, response, nil;
}
