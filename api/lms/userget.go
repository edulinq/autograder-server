package lms

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/canvas"
)

// TEST - Write tests once the LMS abstraction is complete.

type UserGetRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`

    TargetEmail core.NonEmptyString `json:"target-email"`
}

type UserGetResponse struct {
    FoundAGUser bool `json:"found-autograder-user"`
    FoundLMSUser bool `json:"found-lms-user"`
    User *core.UserInfo `json:"user"`
}

func HandleUserGet(request *UserGetRequest) (*UserGetResponse, *core.APIError) {
    if (request.Course.CanvasInstanceInfo == nil) {
        return nil, core.NewBadRequestError("-501", &request.APIRequest, "Course is not linked to an LMS.").
                Add("course", request.Course.ID);
    }

    response := UserGetResponse{};

    user := request.Users[string(request.TargetEmail)];
    if (user == nil) {
        return &response, nil;
    }

    response.FoundAGUser = true;
    response.User = core.NewUserInfo(user);

    lmsUser, err := canvas.FetchUser(request.Course.CanvasInstanceInfo, string(request.TargetEmail));
    if (err != nil) {
        return nil, core.NewInternalError("-502", &request.APIRequestCourseUserContext,
                "Failed to fetch canvas user.").Err(err).Add("email", string(request.TargetEmail));
    }

    if (lmsUser == nil) {
        return &response, nil;
    }

    response.FoundLMSUser = true;
    response.User.Name = lmsUser.Name;
    response.User.LMSID = lmsUser.ID;

    return &response, nil;
}
