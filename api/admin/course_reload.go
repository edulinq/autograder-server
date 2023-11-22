package admin

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
)

type CourseReloadRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin
}

type CourseReloadResponse struct {
}

func HandleCourseReload(request *CourseReloadRequest) (*CourseReloadResponse, *core.APIError) {
    _, err := db.ReloadCourse(request.Course);
    if (err != nil) {
        return nil, core.NewInternalError("-701", &request.APIRequestCourseUserContext,
                "Failed to reload course.").Err(err);
    }

    return &CourseReloadResponse{}, nil;
}
