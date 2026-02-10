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
	ImageInfo *docker.BuiltImageInfo `json:"image-info"`
	Gzip      bool                   `json:"gzip"`
	Bytes     string                 `json:"bytes"`
}

// Fetch an assignment's current Docker image.
func HandleFetch(request *FetchRequest) (*FetchResponse, *core.APIError) {
	// Ensure the assignment docker image is built.
	err := docker.BuildImageFromSourceQuick(request.Assignment)
	if err != nil {
		return nil, core.NewInternalError("-642", request, "Failed to build assignment image.").Err(err)
	}

	imageInfo, err := docker.GetImage(request.Assignment.ImageName(), true)
	if err != nil {
		return nil, core.NewInternalError("-643", request, "Failed to fetch image info.").Err(err)
	}

	response := FetchResponse{
		ImageInfo: imageInfo,
		Gzip:      true,
		Bytes:     util.Base64Encode(imageInfo.GzipBytes),
	}

	return &response, nil
}
