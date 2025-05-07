package metadata

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

type HeartbeatRequest struct {
	core.APIRequest
}

type HeartbeatResponse struct {
	ServerVersion util.Version `json:"server-version"`
}

// Get server heartbeat.
func HandleHeartbeat(request *HeartbeatRequest) (*HeartbeatResponse, *core.APIError) {
	version, err := util.GetFullCachedVersion()
	if err != nil {
		return nil, core.NewInternalError("-502", request, "Unable to get server version.").Err(err)
	}

	response := HeartbeatResponse{
		ServerVersion: version,
	}

	return &response, nil
}
