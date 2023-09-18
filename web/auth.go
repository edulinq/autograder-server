package web

import (
    "fmt"

    "github.com/rs/zerolog/log"

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

    user, err := course.GetUser(request.User);
    if (err != nil) {
        log.Debug().Str("user", request.User).Err(err).Msg("Authentication failure.");
        return false, nil;
    }

    if (user.Pass != request.Pass) {
        return false, nil;
    }

    return true, nil;
}
