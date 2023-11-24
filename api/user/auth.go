package user

import (
    "github.com/eriq-augustine/autograder/api/core"
)

type AuthRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleOther

    TargetUser core.TargetUser `json:"target-email"`
    TargetPass core.NonEmptyString `json:"target-pass"`
}

type AuthResponse struct {
    FoundUser bool `json:"found-user"`
    AuthSuccess bool `json:"auth-success"`
}

func HandleAuth(request *AuthRequest) (*AuthResponse, *core.APIError) {
    response := AuthResponse{};

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;
    response.AuthSuccess = request.TargetUser.User.CheckPassword(string(request.TargetPass));

    return &response, nil;
}
