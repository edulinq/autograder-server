package api

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
)

// Return a user only in the case that the authentication is successful.
// No error is retutned if authentication fails or the user cannot be found, a nil user will just be returned.
// For security not much error information is leaked from this function,
// but more information is put on the debug log.
func AuthAPIRequest(course *model.Course, email string, pass string) (*usr.User, error) {
    if (course == nil) {
        return nil, fmt.Errorf("Cannot authenticate nil course.");
    }

    user, err := course.GetUser(email);
    if (err != nil) {
        log.Debug().Str("email", email).Str("course", course.ID).Err(err).Msg("Authentication Failure: Cannot get user.");
        return nil, err;
    }

    if (user == nil) {
        log.Debug().Str("email", email).Str("course", course.ID).Msg("Authentication Failure: Unknown user.");
        return nil, nil;
    }

    if (config.NO_AUTH.GetBool()) {
        log.Debug().Str("email", email).Str("course", course.ID).Msg("Authentication Disabled.");
        return user, nil;
    }

    if (!user.CheckPassword(pass)) {
        log.Debug().Str("email", email).Str("course", course.ID).Msg("Authentication Failure: Bad Password.");
        return nil, nil;
    }

    return user, nil;
}
