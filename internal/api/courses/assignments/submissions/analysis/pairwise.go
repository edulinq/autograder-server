package analysis

import (
	"fmt"

	"github.com/edulinq/autograder/internal/analysis"
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
)

type PairwiseRequest struct {
	core.APIRequestUserContext
	core.MinServerRoleUser

	analysis.AnalysisOptions
}

type PairwiseResponse struct {
	Complete bool                           `json:"complete"`
	Options  analysis.AnalysisOptions       `json:"options"`
	Summary  *model.PairwiseAnalysisSummary `json:"summary"`
	Results  model.PairwiseAnalysisMap      `json:"results"`
}

// Get the result of a pairwise analysis for the specified submissions.
func HandlePairwise(request *PairwiseRequest) (*PairwiseResponse, *core.APIError) {
	fullSubmissionIDs, courses, userErrors, systemErrors := analysis.ResolveSubmissionSpecs(request.RawSubmissionSpecs)

	if systemErrors != nil {
		return nil, core.NewInternalError("-619", request, "Failed to resolve submission specs.").
			Err(systemErrors)
	}

	if userErrors != nil {
		return nil, core.NewBadRequestError("-620", request,
			fmt.Sprintf("Failed to resolve submission specs: '%s'.", userErrors.Error())).
			Err(userErrors)
	}

	if !checkPermissions(request.ServerUser, courses) {
		return nil, core.NewBadRequestError("-621", request,
			"User does not have permissions (server admin or course admin in all present courses.")
	}

	request.ResolvedSubmissionIDs = fullSubmissionIDs
	request.InitiatorEmail = request.ServerUser.Email
	request.AnalysisOptions.Context = request.APIRequestUserContext.Context

	results, pendingCount, err := analysis.PairwiseAnalysis(request.AnalysisOptions)
	if err != nil {
		return nil, core.NewInternalError("-622", request, "Failed to perform pairwise analysis.").
			Err(err)
	}

	response := PairwiseResponse{
		Complete: (pendingCount == 0),
		Options:  request.AnalysisOptions,
		Summary:  model.NewPairwiseAnalysisSummary(results, pendingCount),
		Results:  results,
	}

	return &response, nil
}
