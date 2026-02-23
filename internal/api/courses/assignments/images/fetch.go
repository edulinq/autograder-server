package images

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/docker"
	"github.com/edulinq/autograder/internal/util"
)

type FetchRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader
}

type FetchResponse struct {
	ImageInfo *docker.BuiltImageInfoAndData `json:"image-info"`
	Gzip      bool                          `json:"gzip"`
	Bytes     string                        `json:"bytes"`
}

// Fetch an assignment's current Docker image.
func HandleFetch(request *FetchRequest) (*FetchResponse, *core.APIError) {
	imageInfo, err := docker.GetBuiltImageInfoAndData(request.Assignment, true)
	if err != nil {
		return nil, core.NewInternalError("-642", request, "Failed to fetch built image info and data.").Err(err)
	}

	response := FetchResponse{
		ImageInfo: imageInfo,
		Gzip:      true,
		Bytes:     util.Base64Encode(imageInfo.GzipBytes),
	}

	return &response, nil
}
