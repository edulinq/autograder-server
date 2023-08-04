package grader

import (
    "fmt"

    "github.com/eriq-augustine/autograder/model"
)

// Return true only if the request is authenticated.
func AuthAPIRequest(request *model.BaseAPIRequest) (bool, error) {
    // TODO: Auth Users

    if (request == nil) {
        return false, fmt.Errorf("Cannot authenticate nil request.");
    }

    course := GetCourse(request.Course);
    if (course == nil) {
        return false, nil;
    }

    return true, nil;
}
