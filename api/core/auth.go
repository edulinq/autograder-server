package core

import (
    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/model"
)

// Return a user only in the case that the authentication is successful.
// If any error is retuturned, then the request should end and the response sent based on the error.
// This assumes basic validation has already been done on the request.
func (this *APIRequestCourseUserContext) Auth() (*model.User, *APIError) {
    user, err := db.GetUser(this.Course, this.UserEmail);
    if (err != nil) {
        return nil, NewAuthBadRequestError("-012", this, "Cannot Get User").Err(err);
    }

    if (user == nil) {
        return nil, NewAuthBadRequestError("-013", this, "Unknown User");
    }

    if (config.NO_AUTH.Get()) {
        log.Debug("Authentication Disabled.", this.Course, log.NewUserAttr(this.UserEmail));
        return user, nil;
    }

    if (!user.CheckPassword(this.UserPass)) {
        return nil, NewAuthBadRequestError("-014", this, "Bad Password");
    }

    return user, nil;
}
