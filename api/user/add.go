package user

import (
    "fmt"
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/lms/lmssync"
    "github.com/eriq-augustine/autograder/model"
)

type AddRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin
    Users core.CourseUsers `json:"-"`

    NewUsers []*core.UserInfoWithPass `json:"new-users"`

    Force bool `json:"force"`
    DryRun bool `json:"dry-run"`
    SkipEmails bool `json:"skip-emails"`
    SkipLMSSync bool `json:"skip-lms-sync"`
}

type AddResponse struct {
    core.SyncUsersInfo
    Errors []AddError `json:"errors"`
    LMSSyncCount int `jons:"lms-sync-count"`
}

type AddError struct {
    Index int `json:"index"`
    Email string `json:"email"`
    Message string `json:"message"`
}

func HandleAdd(request *AddRequest) (*AddResponse, *core.APIError) {
    response := AddResponse{
        Errors: []AddError{},
    };

    newUsers := make(map[string]*model.User, len(request.NewUsers));
    emails := make([]string, 0, len(request.NewUsers));

    for i, apiUser := range request.NewUsers {
        user, err := apiUser.ToUsr();
        if (err != nil) {
            response.Errors = append(response.Errors, AddError{i, apiUser.Email, err.Error()});
            continue;
        }

        if (user.Role > request.User.Role) {
            message := fmt.Sprintf("Cannot create a user with a higher role (%s) than your role (%s).",
                    user.Role.String(), request.User.Role.String());
            response.Errors = append(response.Errors, AddError{i, apiUser.Email, message});
            continue;
        }

        localUser, ok := request.Users[user.Email];
        if (ok && (localUser.Role > request.User.Role)) {
            message := fmt.Sprintf("Cannot modify a user with a higher role (%s) than your role (%s).",
                    localUser.Role.String(), request.User.Role.String());
            response.Errors = append(response.Errors, AddError{i, apiUser.Email, message});
            continue;
        }

        newUsers[apiUser.Email] = user;
        emails = append(emails, apiUser.Email);
    }

    result, err := db.SyncUsers(request.Course, newUsers, request.Force, request.DryRun, !request.SkipEmails);
    if (err != nil) {
        return nil, core.NewInternalError("-803", &request.APIRequestCourseUserContext,
                "Failed to sync new users.").Err(err);
    }

    if (!request.SkipLMSSync) {
        lmsResult, err := lmssync.SyncLMSUserEmails(request.Course, emails, request.DryRun, !request.SkipEmails);
        if (err != nil) {
            log.Error().Err(err).Str("api-request", request.RequestID).Msg("Failed to sync LMS users.");
        } else {
            response.LMSSyncCount = lmsResult.Count();
        }

        // Users may have been updated in the LMS sync.

        users, err := db.GetUsers(request.Course);
        if (err != nil) {
            return nil, core.NewInternalError("-804", &request.APIRequestCourseUserContext,
                    "Failed to fetch users").Err(err);
        }

        result.UpdateUsers(users);
    }

    response.SyncUsersInfo = *core.NewSyncUsersInfo(result);

    return &response, nil;
}
