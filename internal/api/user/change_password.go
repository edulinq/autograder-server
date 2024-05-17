package user

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
)

type ChangePasswordRequest struct {
	core.APIRequestCourseUserContext
	core.MinRoleOther

	TargetUser core.TargetUserSelfOrAdmin `json:"target-email"`
	NewPass    string                     `json:"new-pass"`
}

type ChangePasswordResponse struct {
	FoundUser bool `json:"found-user"`
}

func HandleChangePassword(request *ChangePasswordRequest) (*ChangePasswordResponse, *core.APIError) {
	response := ChangePasswordResponse{}

	if !request.TargetUser.Found {
		return &response, nil
	}

	response.FoundUser = true

	if request.TargetUser.User.Role > request.User.Role {
		return nil, core.NewBadPermissionsError("-805", &request.APIRequestCourseUserContext, request.TargetUser.User.Role,
			"Cannot modify a user with a higher role.").Add("target-user", request.TargetUser.User.Email)
	}

	var err error
	var pass string

	if request.NewPass == "" {
		pass, err = request.TargetUser.User.SetRandomPassword()
	} else {
		err = request.TargetUser.User.SetPassword(request.NewPass)
	}

	if err != nil {
		return nil, core.NewInternalError("-806", &request.APIRequestCourseUserContext,
			"Failed to set password.").Err(err).Add("target-user", request.TargetUser.Email)
	}

	err = db.SaveUser(request.Course, request.TargetUser.User)
	if err != nil {
		return nil, core.NewInternalError("-807", &request.APIRequestCourseUserContext,
			"Failed to save user.").Err(err).Add("target-user", request.TargetUser.Email)
	}

	if pass != "" {
		err = model.SendUserAddEmail(request.Course, request.TargetUser.User, pass, true, true, false, false)
		if err != nil {
			return nil, core.NewInternalError("-808", &request.APIRequestCourseUserContext,
				"Failed to send user email.").Err(err).Add("target-user", request.TargetUser.Email)
		}
	}

	return &response, nil
}
