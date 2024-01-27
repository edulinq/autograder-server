package admin

import (
    "fmt"
    "time"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/log"
)

type FetchLogsRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    Level string `json:"level"`
    After string `json:"after"`

    AssignmentID string `json:"assignment-id"`
    TargetUser string `json:"target-email"`
}

type FetchLogsResponse struct {
    Success bool `json:"success"`
    ErrorMessages []string `json:"error-messages"`
    Records []*log.Record `json:"results"`
}

func HandleFetchLogs(request *FetchLogsRequest) (*FetchLogsResponse, *core.APIError) {
    var response FetchLogsResponse;

    if (request.Level == "") {
        request.Level = log.LEVEL_STRING_INFO;
    }

    level, err := log.ParseLevel(request.Level);
    if (err != nil) {
        response.ErrorMessages = append(response.ErrorMessages, fmt.Sprintf("Could not parse 'level': '%v'.", err));
    }

    after := time.Time{};
    if (request.After != "") {
        timestamp, err := common.TimestampFromString(request.After);
        if (err != nil) {
            response.ErrorMessages = append(response.ErrorMessages, fmt.Sprintf("Could not parse 'after': '%s'.", request.After));
        } else {
            after, err = timestamp.Time();
            if (err != nil) {
                response.ErrorMessages = append(response.ErrorMessages, fmt.Sprintf("Could not extract time from 'after': '%v'.", err));
            }
        }
    }

    assignment := "";
    if (request.AssignmentID != "") {
        assignment, err = common.ValidateID(request.AssignmentID);
        if (err != nil) {
            response.ErrorMessages = append(response.ErrorMessages, fmt.Sprintf("Improperly formatted 'assignment-id': '%v'.", err));
        } else {
            if (!request.Course.HasAssignment(assignment)) {
                response.ErrorMessages = append(response.ErrorMessages, fmt.Sprintf("Unknown assignment: '%s'.", assignment));
            }
        }
    }

    user := "";
    if (request.TargetUser != "") {
        fullUser, err := db.GetUser(request.Course, request.TargetUser);
        if (err != nil) {
            return nil, core.NewInternalError("-205", &request.APIRequestCourseUserContext, "Failed to get target user.").
                    Add("email", request.TargetUser).Err(err);
        }

        if (fullUser == nil) {
            response.ErrorMessages = append(response.ErrorMessages, fmt.Sprintf("Could not find user: '%s'.", request.TargetUser));
        } else {
            user = fullUser.Email;
        }
    }

    if (len(response.ErrorMessages) > 0) {
        return &response, nil;
    }

    response.Records, err = db.GetLogRecords(level, after, request.Course.GetID(), assignment, user);
    if (err != nil) {
        return nil, core.NewInternalError("-206", &request.APIRequestCourseUserContext, "Failed to get log records.").Err(err);
    }

    response.Success = true;
    return &response, nil;
}
