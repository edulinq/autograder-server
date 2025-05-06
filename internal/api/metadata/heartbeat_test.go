package metadata

import (
	"testing"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/util"
)

func TestHeartbeat(test *testing.T) {
	response := core.SendTestAPIRequest(test, "metadata/heartbeat", nil)
	if !response.Success {
		test.Fatalf("Failed to get heartbeat.")
	}

	var heartbeatResponse HeartbeatResponse
	util.MustJSONFromString(util.MustToJSON(response.Content), &heartbeatResponse)

	expected, err := util.GetFullCachedVersion()
	if err != nil {
		test.Fatalf("Failed to get expected server version: %v", err)
		return
	}
	if heartbeatResponse.ServerVersion != expected {
		test.Fatalf("Server version mismatch. Expected %s, got %s", expected, heartbeatResponse.ServerVersion)
	}
}
