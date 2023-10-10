package web

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
)

// Return true only if the request is authenticated.
// A returned user may be nil if authentication is disabled.
func AuthAPIRequest(request *BaseAPIRequest, course *model.Course) (bool, *usr.User, error) {
    if (request == nil) {
        return false, nil, fmt.Errorf("Cannot authenticate nil request.");
    }

    if (course == nil) {
        return false, nil, fmt.Errorf("Cannot authenticate nil course.");
    }

    user, err := course.GetUser(request.User);

    if (config.NO_AUTH.GetBool()) {
        log.Debug().Str("user", request.User).Msg("Authentication Disabled.");
        return true, user, nil;
    }

    if (err != nil) {
        log.Debug().Str("user", request.User).Err(err).Msg("Authentication Failure: Cannot get user.");
        return false, nil, nil;
    }

    if (user == nil) {
        log.Debug().Str("user", request.User).Msg("Authentication Failure: Unknown user.");
        return false, nil, nil;
    }

    if (!user.CheckPassword(request.Pass)) {
        log.Debug().Str("user", request.User).Msg("Authentication Failure: Bad Password.");
        return false, nil, nil;
    }

    return true, user, nil;
}
