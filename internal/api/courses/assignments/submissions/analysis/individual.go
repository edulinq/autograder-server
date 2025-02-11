package analysis

import (
	"fmt"

	"github.com/edulinq/autograder/internal/analysis"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
)

type IndividualRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	SubmissionSpecs   []string `json:"submissions"`
	WaitForCompletion bool     `json:"wait-for-completion"`
}

type IndividualResponse struct {
	Complete bool                             `json:"complete"`
	Summary  *model.IndividualAnalysisSummary `json:"summary"`
	Results  []*model.IndividualAnalysis      `json:"results"`
}

// Get the result of a individual analysis for the specified submissions.
func HandleIndividual(request *IndividualRequest) (*IndividualResponse, *core.APIError) {
	fullSubmissionIDs, courses, userErrors, systemErrors := analysis.ResolveSubmissionSpecs(request.SubmissionSpecs)

	if systemErrors != nil {
		return nil, core.NewUserContextInternalError("-623", &request.APIRequestUserContext, "Failed to resolve submission specs.").
			Err(systemErrors)
	}

	if userErrors != nil {
		return nil, core.NewBadUserRequestError("-624", &request.APIRequestUserContext,
			fmt.Sprintf("Failed to resolve submission specs: '%s'.", userErrors.Error())).
			Err(userErrors)
	}

	if !checkPermissions(request.ServerUser, courses) {
		return nil, core.NewBadUserRequestError("-625", &request.APIRequestUserContext,
			"User does not have permissions (server admin or course admin in all present courses.")
	}

	results, pendingCount, err := analysis.IndividualAnalysis(fullSubmissionIDs, request.WaitForCompletion, request.ServerUser.Email)
	if err != nil {
		return nil, core.NewUserContextInternalError("-626", &request.APIRequestUserContext, "Failed to perform individual analysis.").
			Err(err)
	}

	response := IndividualResponse{
		Complete: (pendingCount == 0),
		Summary:  model.NewIndividualAnalysisSummary(results, pendingCount),
		Results:  results,
	}

	return &response, nil
}
