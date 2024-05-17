package lms

import (
    "github.com/edulinq/autograder/internal/api/core"
    "github.com/edulinq/autograder/internal/lms/lmssync"
    "github.com/edulinq/autograder/internal/model"
)

type SyncRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    DryRun bool `json:"dry-run"`
    SkipEmails bool `json:"skip-emails"`
}

type SyncResponse struct {
    SyncAvailable bool `json:"sync-available"`
    Users *core.SyncUsersInfo `json:"users"`
    Assignments *model.AssignmentSyncResult `json:"assignments"`
}

func HandleSync(request *SyncRequest) (*SyncResponse, *core.APIError) {
    if (request.Course.GetLMSAdapter() == nil) {
        return nil, core.NewBadRequestError("-403", &request.APIRequest, "Course is not linked to an LMS.").
                Course(request.Course.GetID());
    }

    var response SyncResponse;

    result, err := lmssync.SyncLMS(request.Course, request.DryRun, !request.SkipEmails);
    if (err != nil) {
        return nil, core.NewInternalError("-404", &request.APIRequestCourseUserContext,
                "Failed to sync LMS information.").Err(err);
    }

    if (result == nil) {
        return &response, nil;
    }

    response.SyncAvailable = true;
    response.Users = core.NewSyncUsersInfo(result.UserSync);
    response.Assignments = result.AssignmentSync;

    return &response, nil;
}
