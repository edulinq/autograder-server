package admin

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/procedures/courses"
)

type UpdateRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	Source string `json:"source"`
	Clear  bool   `json:"clear"`
}

type UpdateResponse struct {
	CourseUpdated bool `json:"course-updated"`
}

func HandleUpdate(request *UpdateRequest) (*UpdateResponse, *core.APIError) {
	if request.Clear {
		err := db.ClearCourse(request.Course)
		if err != nil {
			return nil, core.NewInternalError("-608", &request.APIRequestCourseUserContext,
				"Failed to clear course.").Err(err)
		}
	}

	if request.Source != "" {
		spec, err := common.ParseFileSpec(request.Source)
		if err != nil {
			return nil, core.NewBadCourseRequestError("-609", &request.APIRequestCourseUserContext,
				"Source FileSpec is not formatted properly.").Err(err)
		}

		request.Course.Source = spec

		err = db.SaveCourse(request.Course)
		if err != nil {
			return nil, core.NewInternalError("-610", &request.APIRequestCourseUserContext,
				"Failed to save course.").Err(err)
		}
	}

	updated, err := courses.UpdateCourse(request.Course, true)
	if err != nil {
		return nil, core.NewInternalError("-611", &request.APIRequestCourseUserContext,
			"Failed to update course.").Err(err)
	}

	return &UpdateResponse{updated}, nil
}
