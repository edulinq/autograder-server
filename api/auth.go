package api

import (
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/usr"
)

// Return a user only in the case that the authentication is successful.
// If any error is retuturned, then the request should end and the response sent based on the error.
// This assumes basic validation has already been done on the request.
func (this *APIRequestCourseUserContext) Auth() (*usr.User, *APIError) {
    user, err := this.course.GetUser(this.UserEmail);
    if (err != nil) {
        return nil, NewBadAuthError(this, "Cannot Get User").Err(err);
    }

    if (user == nil) {
        return nil, NewBadAuthError(this, "Unknown User");
    }

    if (config.NO_AUTH.GetBool()) {
        log.Debug().Str("email", this.UserEmail).Str("course", this.CourseID).Msg("Authentication Disabled.");
        return user, nil;
    }

    if (!user.CheckPassword(this.UserPass)) {
        return nil, NewBadAuthError(this, "Bad Password");
    }

    return user, nil;
}
