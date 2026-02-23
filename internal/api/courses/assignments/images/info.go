package images

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/docker"
)

type InfoRequest struct {
	core.APIRequestAssignmentContext
	core.MinCourseRoleGrader
}

type InfoResponse struct {
	ImageInfo *docker.BuiltImageInfo `json:"image-info"`
}

// Get information about an assignment's current Docker image (does not include the image itself).
func HandleInfo(request *InfoRequest) (*InfoResponse, *core.APIError) {
	imageInfo, err := docker.GetBuiltImageInfo(request.Assignment, true)
	if err != nil {
		return nil, core.NewInternalError("-643", request, "Failed to fetch built image info.").Err(err)
	}

	response := InfoResponse{
		ImageInfo: imageInfo,
	}

	return &response, nil
}
