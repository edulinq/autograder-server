package lms

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/procedures"
)

type SyncRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    DryRun bool `json:"dry-run"`
    SkipEmails bool `json:"skip-emails"`
}

type SyncResponse struct {
    Users *core.SyncUsersInfo `json:"users"`
    Assignments *model.AssignmentSyncResult `json:"assignments"`
}

func HandleSync(request *SyncRequest) (*SyncResponse, *core.APIError) {
    if (request.Course.GetLMSAdapter() == nil) {
        return nil, core.NewBadRequestError("-403", &request.APIRequest, "Course is not linked to an LMS.").
                Add("course", request.Course.GetID());
    }

    result, err := procedures.SyncLMS(request.Course, request.DryRun, !request.SkipEmails);
    if (err != nil) {
        return nil, core.NewInternalError("-404", &request.APIRequestCourseUserContext,
                "Failed to sync LMS information.").Err(err);
    }

    response := SyncResponse{
        Users: core.NewSyncUsersInfo(result.UserSync),
        Assignments: result.AssignmentSync,
    };

    return &response, nil;
}
