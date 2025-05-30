package upsert

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/procedures/courses"
	"github.com/edulinq/autograder/internal/util"
)

type FileSpecRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleCourseCreator

	courses.CourseUpsertOptions
	FileSpec util.FileSpec `json:"filespec" required:""`
}

// Upsert a course using a filespec.
func HandleFileSpec(request *FileSpecRequest) (*UpsertResponse, *core.APIError) {
	options := request.CourseUpsertOptions
	options.ContextUser = request.ServerUser

	results, err := courses.UpsertFromFileSpec(&request.FileSpec, options)
	if err != nil {
		return nil, core.NewBadRequestError("-614", request,
			"Failed to upsert course from FileSpec.").Err(err)
	}

	return &UpsertResponse{results}, nil
}
