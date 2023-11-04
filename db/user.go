package db

import (
    "fmt"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/usr"
)

func GetCourseUsers(rawCourseID string) (map[string]*usr.User, error) {
    courseID, err := common.ValidateID(rawCourseID);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to validate course id '%s': '%w'.", rawCourseID, err);
    }

    return backend.GetCourseUsers(courseID);
}
