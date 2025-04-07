package logs

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	plogs "github.com/edulinq/autograder/internal/procedures/logs"
)

type QueryRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	log.RawLogQuery
}

type QueryResponse struct {
	Success bool                          `json:"success"`
	Error   *model.ExternalLocatableError `json:"error"`
	Records []*log.Record                 `json:"results"`
}

// Query log entries from the autograder server.
func HandleQuery(request *QueryRequest) (*QueryResponse, *core.APIError) {
	var response QueryResponse

	records, locatableErr, err := plogs.Query(request.RawLogQuery, request.ServerUser)
	if err != nil {
		return nil, core.NewInternalError("-200", request, "Failed to query logs.").Err(err)
	}

	if locatableErr != nil {
		response.Error = locatableErr.ToExternalError()
		return &response, nil
	}

	response.Success = true
	response.Records = records

	return &response, nil
}
