package web

import (
    "fmt"
    "net/http"
    "strconv"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/usr"
)

var MIN_ROLE_LMS_SYNC_USERS usr.UserRole = usr.Admin;
var MIN_ROLE_LMS_UPLOAD_SCORES usr.UserRole = usr.Grader;

// Requests

type LMSSyncUsersRequest struct {
    BaseAPIRequest
}

type LMSScoresUploadRequest struct {
    BaseAPIRequest
    AssignmentLMSID string `json:"assignment-id"`
    Scores [][]string `json:"scores"`
}

// Responses

type LMSSyncUsersResponse struct {
    Success bool `json:"success"`
    Count int `json:"count"`
}

type LMSScoresUploadResponse struct {
    Success bool `json:"success"`
    Count int `json:"count"`
    ErrorCount int `json:"error-count"`
    UnrecognizedUsers []string `json:"unrecognized-users"`
    NoLMSIDUsers []string `json:"no-lms-id-users"`
    BadScores []string `json:"bad-scores"`
}

// Constructors

func NewLMSSyncUsersRequest(request *http.Request) (*LMSSyncUsersRequest, *APIResponse, error) {
    var apiRequest LMSSyncUsersRequest;
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

    if ((user != nil) && (user.Role < MIN_ROLE_LMS_SYNC_USERS)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    if (course.GetLMSAdapter() == nil) {
        return nil, nil, fmt.Errorf("Course '%s' does not have any LMS instance information.", apiRequest.Course);
    }

    return &apiRequest, nil, nil;
}

func NewLMSScoresUploadRequest(request *http.Request) (*LMSScoresUploadRequest, *APIResponse, error) {
    var apiRequest LMSScoresUploadRequest;
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

    if ((user != nil) && (user.Role < MIN_ROLE_LMS_UPLOAD_SCORES)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    if (course.GetLMSAdapter() == nil) {
        return nil, nil, fmt.Errorf("Course '%s' does not have any LMS instance information.", apiRequest.Course);
    }

    return &apiRequest, nil, nil;
}

// Close and Clean

func (this *LMSSyncUsersRequest) Close() error {
    return nil;
}

func (this *LMSSyncUsersRequest) Clean() error {
    return nil;
}

func (this *LMSScoresUploadRequest) Close() error {
    return nil;
}

func (this *LMSScoresUploadRequest) Clean() error {
    return nil;
}

// Handlers

func handleLMSSyncUsers(request *LMSSyncUsersRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    if (course.GetLMSAdapter() == nil) {
        return 0, nil, fmt.Errorf("Course '%s' has no LMS adapter.", request.Course);
    }

    result, err := course.SyncLMSUsers(false, true);
    if (err != nil) {
        return 0, nil, err;
    }

    response := &LMSSyncUsersResponse{
        Success: true,
        Count: result.Count(),
    };

    return 0, response, nil;
}

func handleLMSScoresUpload(request *LMSScoresUploadRequest) (int, any, error) {
    course := grader.GetCourse(request.Course);
    if (course == nil) {
        return 0, nil, fmt.Errorf("Failed to find course '%s'.", request.Course);
    }

    users, err := course.GetUsers();
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to fetch autograder users: '%w'.", err);
    }

    unknownUsers := make([]string, 0);
    noLMSIDUsers := make([]string, 0);
    badScores := make([]string, 0);

    grades := make([]*lms.SubmissionScore, 0, len(request.Scores));

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

        if (user.LMSID == "") {
            noLMSIDUsers = append(noLMSIDUsers, email);
            continue;
        }

        grades = append(grades, &lms.SubmissionScore{
            UserID: user.LMSID,
            Score: score,
        });
    }

    err = course.GetLMSAdapter().UpdateAssignmentScores(request.AssignmentLMSID, grades);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to upload grades: '%w'.", err);
    }

    response := &LMSScoresUploadResponse{
        Success: true,
        Count: len(grades),
        ErrorCount: (len(unknownUsers) + len(noLMSIDUsers) + len(badScores)),
        UnrecognizedUsers: unknownUsers,
        NoLMSIDUsers: noLMSIDUsers,
        BadScores: badScores,
    };

    return 0, response, nil;
}
