package upsert

import (
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/procedures/courses"
)

type ZipFileRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleCourseCreator

	courses.CourseUpsertOptions
	Files core.POSTFiles `json:"-"`
}

// Upsert a course using a zip file.
func HandleZipFile(request *ZipFileRequest) (*UpsertResponse, *core.APIError) {
	if len(request.Files.Filenames) != 1 {
		return nil, core.NewBadUserRequestError("-615", &request.APIRequestUserContext,
			fmt.Sprintf("Expected exactly one file, found %d.", len(request.Files.Filenames)))
	}

	path := filepath.Join(request.Files.TempDir, request.Files.Filenames[0])

	options := request.CourseUpsertOptions
	options.ContextUser = request.ServerUser

	results, err := courses.UpsertFromZipFile(path, options)
	if err != nil {
		return nil, core.NewBadUserRequestError("-616", &request.APIRequestUserContext,
			"Failed to upsert course from zip file.").Err(err)
	}

	return &UpsertResponse{results}, nil
}
