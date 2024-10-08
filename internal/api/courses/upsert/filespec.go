package upsert

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/procedures/courses"
)

type FileSpecRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleCourseCreator

	courses.CourseUpsertOptions
	FileSpec common.FileSpec `json:"filespec"`
}

func HandleFileSpec(request *FileSpecRequest) (*UpsertResponse, *core.APIError) {
	options := request.CourseUpsertOptions
	options.ContextUser = request.ServerUser

	results, err := courses.UpsertFromFileSpec(&request.FileSpec, options)
	if err != nil {
		return nil, core.NewBadUserRequestError("-614", &request.APIRequestUserContext,
			"Failed to upsert course from FileSpec.").Err(err)
	}

	return &UpsertResponse{results}, nil
}
