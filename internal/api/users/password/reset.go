package password

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/model"
)

type PasswordResetRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	// If set, reset this user's password instead of the caller's own.
	// Only server admins may specify a target other than themselves.
	TargetUserEmail string `json:"target-user-email"`
}

type PasswordResetResponse struct{}

// Reset to a random password that will be emailed to you.
func HandlePasswordReset(request *PasswordResetRequest) (*PasswordResetResponse, *core.APIError) {
	response := &PasswordResetResponse{}

	targetEmail := request.ServerUser.Email

	if request.TargetUserEmail != "" {
		if request.ServerUser.Role < model.ServerRoleAdmin {
			return nil, core.NewPermissionsError("-811", request, model.ServerRoleAdmin, request.ServerUser.Role, "Resetting another user's password requires a server admin role.")
		}

		targetEmail = request.TargetUserEmail
	}

	user, err := db.GetServerUser(targetEmail)
	if err != nil {
		return nil, core.NewInternalError("-807", request, "Failed to get server user.").Err(err)
	}

	if user == nil {
		return response, nil
	}

	cleartext, err := user.SetRandomPassword()
	if err != nil {
		return nil, core.NewInternalError("-808", request, "Failed to set random password.").Err(err)
	}

	err = db.UpsertUser(user)
	if err != nil {
		return nil, core.NewInternalError("-809", request, "Failed to save user.").Err(err)
	}

	userOp := &model.UserOpResult{
		BaseUserOpResult: model.BaseUserOpResult{
			Email:    targetEmail,
			Modified: true,
		},
		CleartextPassword: cleartext,
	}

	message := userOp.GetEmail()
	if message == nil {
		return response, nil
	}

	err = email.SendMessage(message)
	if err != nil {
		return nil, core.NewInternalError("-810", request, "Failed to email user.").Err(err)
	}

	return response, nil
}
