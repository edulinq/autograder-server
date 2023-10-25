package lms

import (
    "github.com/eriq-augustine/autograder/api/core"
)

type SyncUsersRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    DryRun bool `json:"dry-run"`
    SendEmails bool `json:"send-emails"`
}

type SyncUsersResponse struct {
    Add []*core.UserInfo `json:"add-users"`
    Mod []*core.UserInfo `json:"mod-users"`
    Del []*core.UserInfo `json:"del-users"`
    Skip []*core.UserInfo `json:"skip-users"`
}

func HandleSyncUsers(request *SyncUsersRequest) (*SyncUsersResponse, *core.APIError) {
    if (request.Course.LMSAdapter == nil) {
        return nil, core.NewBadRequestError("-503", &request.APIRequest, "Course is not linked to an LMS.").
                Add("course", request.Course.ID);
    }

    result, err := request.Course.SyncLMSUsers(request.DryRun, request.SendEmails);
    if (err != nil) {
        return nil, core.NewInternalError("-504", &request.APIRequestCourseUserContext,
                "Failed to sync LMS users.").Err(err);
    }

    response := SyncUsersResponse{
        Add: core.NewUserInfos(result.Add),
        Mod: core.NewUserInfos(result.Mod),
        Del: core.NewUserInfos(result.Del),
        Skip: core.NewUserInfos(result.Skip),
    };

    return &response, nil;
}
