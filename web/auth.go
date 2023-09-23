package web

import (
    "fmt"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/model"
)

// Return true only if the request is authenticated.
func AuthAPIRequest(request *model.BaseAPIRequest, course *model.Course) (bool, error) {
    if (request == nil) {
        return false, fmt.Errorf("Cannot authenticate nil request.");
    }

    if (course == nil) {
        return false, fmt.Errorf("Cannot authenticate nil course.");
    }

    if (config.NO_AUTH.GetBool()) {
        log.Debug().Str("user", request.User).Msg("Authentication Disabled.");
        return true, nil;
    }

    user, err := course.GetUser(request.User);
    if (err != nil) {
        log.Debug().Str("user", request.User).Err(err).Msg("Authentication Failure: Unknown User.");
        return false, nil;
    }

    if (!user.CheckPassword(request.Pass)) {
        log.Debug().Str("user", request.User).Msg("Authentication Failure: Bad Password.");
        return false, nil;
    }

    return true, nil;
}
