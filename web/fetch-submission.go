package web

import (
    "fmt"
    "net/http"
    "path/filepath"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

var MIN_ROLE_FETCH_SUBMISSION usr.UserRole = usr.Grader;

type FetchSubmissionRequest struct {
    BaseAPIRequest
    Assignment string `json:"assignment"`
    Email string `json:"email"`
}

type FetchSubmissionResponse struct {
    FoundUser bool `json:"found-user"`
    FoundSubmission bool `json:"found-submission"`
    SubmissionID string `json:"submission-id"`
    Contents string `json:"contents"`
}

func (this *FetchSubmissionRequest) String() string {
    return util.BaseString(this);
}

func NewFetchSubmissionRequest(request *http.Request) (*FetchSubmissionRequest, *APIResponse, error) {
    var apiRequest FetchSubmissionRequest;
    err := APIRequestFromPOST(&apiRequest, request);
    if (err != nil) {
        return nil, nil, err;
    }

    course, _, err := grader.VerifyCourseAssignment(apiRequest.Course, apiRequest.Assignment);
    if (err != nil) {
        return nil, nil, err;
    }

    ok, user, err := AuthAPIRequest(&apiRequest.BaseAPIRequest, course);
    if (err != nil) {
        return nil, nil, err;
    } else if (!ok) {
        return nil, NewResponse(http.StatusUnauthorized, "Failed to authenticate."), nil;
    }

    if ((user != nil) && (user.Role < MIN_ROLE_FETCH_SUBMISSION)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    return &apiRequest, nil, nil;
}

func (this *FetchSubmissionRequest) Close() error {
    return nil;
}

func (this *FetchSubmissionRequest) Clean() error {
    var err error;
    this.Assignment, err = common.ValidateID(this.Assignment);
    if (err != nil) {
        return fmt.Errorf("Could not clean FetchSubmissionRequest assignment ID ('%s'): '%w'.", this.Assignment, err);
    }

    return nil;
}

func handleFetchSubmission(request *FetchSubmissionRequest) (int, any, error) {
    assignment := grader.GetAssignment(request.Course, request.Assignment);
    if (assignment == nil) {
        return http.StatusBadRequest, fmt.Sprintf("Could not find assignment ('%s') for course ('%s').", request.Assignment, request.Course,), nil;
    }

    users, err := assignment.Course.GetUsers();
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to get users for course: '%w'.", err);
    }

    response := FetchSubmissionResponse{};

    user, ok := users[request.Email];
    if (!ok) {
        return 0, response, nil;
    }

    response.FoundUser = true;

    paths, err := assignment.GetSubmissionSummaries(user.Email);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to get user '%s' submission: '%w'.", user.Email, err);
    }

    if (len(paths) == 0) {
        return 0, response, nil;
    }
    path := paths[len(paths) - 1];

    response.FoundSubmission = true;

    summary := artifact.SubmissionSummary{};
    err = util.JSONFromFile(path, &summary);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to deserialize submission summary '%s': '%w'.", path, err);
    }

    response.SubmissionID = summary.ID;

    submissionPath := filepath.Dir(filepath.Dir(path));
    data, err := util.ZipToBytes(submissionPath, response.SubmissionID);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to create a zip file in memory for submission '%s': '%w'.", submissionPath, err);
    }

    response.Contents = util.Base64Encode(data);

    return 0, response, nil;
}
