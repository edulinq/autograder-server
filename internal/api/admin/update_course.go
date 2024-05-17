package admin

import (
    "github.com/edulinq/autograder/internal/api/core"
    "github.com/edulinq/autograder/internal/common"
    "github.com/edulinq/autograder/internal/db"
    "github.com/edulinq/autograder/internal/procedures"
)

type UpdateCourseRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleAdmin

    Source string `json:"source"`
    Clear bool `json:"clear"`
}

type UpdateCourseResponse struct {
    CourseUpdated bool `json:"course-updated"`
}

func HandleUpdateCourse(request *UpdateCourseRequest) (*UpdateCourseResponse, *core.APIError) {
    if (request.Clear) {
        err := db.ClearCourse(request.Course);
        if (err != nil) {
            return nil, core.NewInternalError("-201", &request.APIRequestCourseUserContext,
                    "Failed to clear course.").Err(err);
        }
    }

    if (request.Source != "") {
        spec, err := common.ParseFileSpec(request.Source);
        if (err != nil) {
            return nil, core.NewBadCourseRequestError("-202", &request.APIRequestCourseUserContext,
                    "Source FileSpec is not formatted properly.").Err(err);
        }

        request.Course.Source = spec;

        err = db.SaveCourse(request.Course);
        if (err != nil) {
            return nil, core.NewInternalError("-203", &request.APIRequestCourseUserContext,
                    "Failed to save course.").Err(err);
        }
    }

    updated, err := procedures.UpdateCourse(request.Course, true);
    if (err != nil) {
        return nil, core.NewInternalError("-204", &request.APIRequestCourseUserContext,
                "Failed to update course.").Err(err);
    }

    return &UpdateCourseResponse{updated}, nil;
}
