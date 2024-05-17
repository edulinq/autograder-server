package admin

import (
    "fmt"

    "github.com/edulinq/autograder/internal/api/core"
    "github.com/edulinq/autograder/internal/common"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/log"
)

type FetchLogsRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    common.RawLogQuery
}

type FetchLogsResponse struct {
    Success bool `json:"success"`
    ErrorMessages []string `json:"error-messages"`
    Records []*log.Record `json:"results"`
}

func HandleFetchLogs(request *FetchLogsRequest) (*FetchLogsResponse, *core.APIError) {
    var response FetchLogsResponse;

    parsedQuery, messages := request.RawLogQuery.ParseStrings(request.Course);
    if (len(messages) > 0) {
        response.ErrorMessages = messages;
        return &response, nil;
    }

    if (parsedQuery.UserID != "") {
        fullUser, err := db.GetUser(request.Course, parsedQuery.UserID);
        if (err != nil) {
            return nil, core.NewInternalError("-205", &request.APIRequestCourseUserContext, "Failed to get target user.").
                    Add("target-user", parsedQuery.UserID).Err(err);
        }

        if (fullUser == nil) {
            response.ErrorMessages = append(response.ErrorMessages, fmt.Sprintf("Could not find user: '%s'.", parsedQuery.UserID));
        } else {
            parsedQuery.UserID = fullUser.Email;
        }
    }

    if (len(response.ErrorMessages) > 0) {
        return &response, nil;
    }

    var err error;
    response.Records, err = db.GetLogRecords(parsedQuery.Level, parsedQuery.After,
            request.Course.GetID(), parsedQuery.AssignmentID, parsedQuery.UserID);
    if (err != nil) {
        return nil, core.NewInternalError("-206", &request.APIRequestCourseUserContext, "Failed to get log records.").Err(err);
    }

    response.Success = true;
    return &response, nil;
}
