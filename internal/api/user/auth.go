package user

import (
    "github.com/edulinq/autograder/api/core"
)

type AuthRequest struct {
    core.APIRequestCourseUserContext
    core.MinCourseRoleOther

    TargetPass core.NonEmptyString `json:"target-pass"`
}

type AuthResponse struct {
    // FoundUser bool `json:"found-user"`
    AuthSuccess bool `json:"auth-success"`
}

func HandleAuth(request *AuthRequest) (*AuthResponse, *core.APIError) {
    response := AuthResponse{};

    // if (!request.TargetUser.Found) {
    //     return &response, nil;
    // }

    // response.FoundUser = true;
	user, err := db.GetServerUser(this.UserEmail, false)
	if (err != nil) {
        return nil, NewAuthBadRequestError("-012", this, "Cannot Get User").Err(err);
    }

    if (user == nil) {
        return nil, NewAuthBadRequestError("-013", this, "Uknown User");
    }

    response.AuthSuccess = user.CheckPassword(string(request.TargetPass));

    return &response, nil;
}
