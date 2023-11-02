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

var MIN_ROLE_FETCH_SUBMISSIONS usr.UserRole = usr.Grader;

type FetchSubmissionsRequest struct {
    BaseAPIRequest
    Assignment string `json:"assignment"`
}

type FetchSubmissionsResponse struct {
    SubmissionIDs map[string]string `json:"submission-ids"`
    Contents string `json:"contents"`
}

func NewFetchSubmissionsRequest(request *http.Request) (*FetchSubmissionsRequest, *APIResponse, error) {
    var apiRequest FetchSubmissionsRequest;
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

    if ((user != nil) && (user.Role < MIN_ROLE_FETCH_SUBMISSIONS)) {
        log.Debug().Str("user", user.Email).Msg("Authentication Failure: Insufficient Permissions.");
        return nil, NewResponse(http.StatusForbidden, "Insufficient Permissions."), nil;
    }

    return &apiRequest, nil, nil;
}

func (this *FetchSubmissionsRequest) Close() error {
    return nil;
}

func (this *FetchSubmissionsRequest) Clean() error {
    var err error;
    this.Assignment, err = common.ValidateID(this.Assignment);
    if (err != nil) {
        return fmt.Errorf("Could not clean FetchSubmissionsRequest assignment ID ('%s'): '%w'.", this.Assignment, err);
    }

    return nil;
}

func handleFetchSubmissions(request *FetchSubmissionsRequest) (int, any, error) {
    assignment := grader.GetAssignment(request.Course, request.Assignment);
    if (assignment == nil) {
        return http.StatusBadRequest, fmt.Sprintf("Could not find assignment ('%s') for course ('%s').", request.Assignment, request.Course), nil;
    }

    users, err := assignment.GetCourse().GetUsers();
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to get users for course: '%w'.", err);
    }

    response := FetchSubmissionsResponse{
        SubmissionIDs: make(map[string]string, len(users)),
    };

    paths, err := assignment.GetAllRecentSubmissionSummaries(users);
    if (err != nil) {
        return 0, nil, fmt.Errorf("Failed to get submissions: '%w'.", err);
    }

    zipOperation := util.NewOngoingZipOperation(false);
    defer zipOperation.Close();

    for email, path := range paths {
        if (path == "") {
            continue;
        }

        summary := artifact.SubmissionSummary{};
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return 0, nil, fmt.Errorf("Failed to deserialize submission summary '%s': '%w'.", path, err);
        }

        submissionPath := filepath.Dir(filepath.Dir(path));
        submissionID := common.GetShortSubmissionID(summary.ID);

        err = zipOperation.AddDir(submissionPath, filepath.Join("submissions", email));
        if (err != nil) {
            return 0, nil, fmt.Errorf("Failed to add submission dir ('%s') to zip: '%w'.", submissionPath, err);
        }

        response.SubmissionIDs[email] = submissionID;
    }

    response.Contents = util.Base64Encode(zipOperation.GetBytes());

    return 0, response, nil;
}
