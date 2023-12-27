package lms

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/procedures"
)

type SyncUsersRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    DryRun bool `json:"dry-run"`
    SkipEmails bool `json:"skip-emails"`
}

type SyncUsersResponse struct {
    core.SyncUsersInfo
}

func HandleSyncUsers(request *SyncUsersRequest) (*SyncUsersResponse, *core.APIError) {
    if (request.Course.GetLMSAdapter() == nil) {
        return nil, core.NewBadRequestError("-503", &request.APIRequest, "Course is not linked to an LMS.").
                Add("course", request.Course.GetID());
    }

    result, err := procedures.SyncAllLMSUsers(request.Course, request.DryRun, !request.SkipEmails);
    if (err != nil) {
        return nil, core.NewInternalError("-504", &request.APIRequestCourseUserContext,
                "Failed to sync LMS users.").Err(err);
    }

    response := SyncUsersResponse{
        SyncUsersInfo: *core.NewSyncUsersInfo(result),
    };

    return &response, nil;
}
