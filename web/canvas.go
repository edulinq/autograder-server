package web

import (
    "fmt"
    "net/http"
    "strconv"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/canvas"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/usr"
)

var MIN_ROLE_CANVAS_SYNC_USERS usr.UserRole = usr.Admin;
var MIN_ROLE_CANVAS_UPLOAD_SCORES usr.UserRole = usr.Grader;

// Requests

type CanvasSyncUsersRequest struct {
    BaseAPIRequest
}

type CanvasScoresUploadRequest struct {
    BaseAPIRequest
    AssignmentCanvasID string `json:"assignment-id"`
    Scores [][]string `json:"scores"`
}

// Responses

type CanvasSyncUsersResponse struct {
    Success bool `json:"success"`
    Count int `json:"count"`
}

type CanvasScoresUploadResponse struct {
    Success bool `json:"success"`
    Count int `json:"count"`
    ErrorCount int `json:"error-count"`
    UnrecognizedUsers []string `json:"unrecognized-users"`
    NoCanvasIDUsers []string `json:"no-canvas-id-users"`
    BadScores []string `json:"bad-scores"`
}

// Constructors

func NewCanvasSyncUsersRequest(request *http.Request) (*CanvasSyncUsersRequest, *APIResponse, error) {
    var apiRequest CanvasSyncUsersRequest;
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

    if ((user != nil) && (user.Role < MIN_ROLE_CANVAS_SYNC_USERS)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    if (course.CanvasInstanceInfo == nil) {
        return nil, nil, fmt.Errorf("Course '%s' does not have any canvas instance information.", apiRequest.Course);
    }

    return &apiRequest, nil, nil;
}

func NewCanvasScoresUploadRequest(request *http.Request) (*CanvasScoresUploadRequest, *APIResponse, error) {
    var apiRequest CanvasScoresUploadRequest;
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

    if ((user != nil) && (user.Role < MIN_ROLE_CANVAS_UPLOAD_SCORES)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
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

func (this *CanvasScoresUploadRequest) Close() error {
    return nil;
}

func (this *CanvasScoresUploadRequest) Clean() error {
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

func handleCanvasScoresUpload(request *CanvasScoresUploadRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    users, err := course.GetUsers();
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to fetch autograder users: '%w'.", err);
    }

    unknownUsers := make([]string, 0);
    noCanvasIDUsers := make([]string, 0);
    badScores := make([]string, 0);

    grades := make([]*canvas.CanvasGradeInfo, 0, len(request.Scores));

    for i, row := range request.Scores {
        if (len(row) != 2) {
            return 0, nil, fmt.Errorf("Score row %d has the wrong number of elements. Has %d, expected 2.", i, len(row));
        }

        email := row[0];

        score, err := strconv.ParseFloat(row[1], 64);
        if (err != nil) {
            badScores = append(badScores, row[1]);
            continue;
        }

        user := users[email];
        if (user == nil) {
            unknownUsers = append(unknownUsers, email);
            continue;
        }

        if (user.CanvasID == "") {
            noCanvasIDUsers = append(noCanvasIDUsers, email);
            continue;
        }

        grades = append(grades, &canvas.CanvasGradeInfo{
            UserID: user.CanvasID,
            Score: score,
        });
    }

    err = canvas.UpdateAssignmentGrades(course.CanvasInstanceInfo, request.AssignmentCanvasID, grades);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to upload grades: '%w'.", err);
    }

    response := &CanvasScoresUploadResponse{
        Success: true,
        Count: len(grades),
        ErrorCount: (len(unknownUsers) + len(noCanvasIDUsers) + len(badScores)),
        UnrecognizedUsers: unknownUsers,
        NoCanvasIDUsers: noCanvasIDUsers,
        BadScores: badScores,
    };

    return 0, response, nil;
}
