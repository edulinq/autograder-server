package admin

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
)

type CourseReloadRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    Clear bool `json:"clear"`
}

type CourseReloadResponse struct {
}

func HandleCourseReload(request *CourseReloadRequest) (*CourseReloadResponse, *core.APIError) {
    if (request.Clear) {
        err := db.ClearCourse(request.Course);
        if (err != nil) {
            return nil, core.NewInternalError("-701", &request.APIRequestCourseUserContext,
                    "Failed to clear course.").Err(err);
        }
    }

    _, err := db.UpdateCourseFromSource(request.Course);
    if (err != nil) {
        return nil, core.NewInternalError("-702", &request.APIRequestCourseUserContext,
                "Failed to reload course.").Err(err);
    }

    return &CourseReloadResponse{}, nil;
}
