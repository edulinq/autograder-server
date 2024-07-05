package core

import (
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// Return a user only in the case that the authentication is successful.
// If any error is retuturned, then the request should end and the response sent based on the error.
// This assumes basic validation has already been done on the request.
func (this *APIRequestUserContext) Auth() (*model.ServerUser, *APIError) {
	user, err := db.GetServerUser(this.UserEmail, true)
	if err != nil {
		return nil, NewAuthBadRequestError("-012", this, "Cannot Get User").Err(err)
	}

	if user == nil {
		return nil, NewAuthBadRequestError("-013", this, "Unknown User")
	}

	if config.NO_AUTH.Get() {
		log.Debug("Authentication Disabled.", log.NewUserAttr(this.UserEmail))
		return user, nil
	}

	auth, err := user.Auth(this.UserPass)
	if err != nil {
		return nil, NewBareInternalError("-037", this.Endpoint, "User auth failed.").Err(err)
	}

	if !auth {
		return nil, NewAuthBadRequestError("-014", this, "Bad Password")
	}

	return user, nil
}
