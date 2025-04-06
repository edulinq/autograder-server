package password

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/model"
)

type PasswordResetRequest struct {
	core.APIRequest

	UserEmail core.NonEmptyString `json:"user-email"`
}

type PasswordResetResponse struct{}

// Reset to a random password that will be emailed to you.
func HandlePasswordReset(request *PasswordResetRequest) (*PasswordResetResponse, *core.APIError) {
	response := &PasswordResetResponse{}

	user, err := db.GetServerUser(string(request.UserEmail))
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
			Email:    string(request.UserEmail),
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
