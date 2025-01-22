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

	SubmissionSpecs   []string `json:"submissions"`
	WaitForCompletion bool     `json:"wait-for-completion"`
}

type PairwiseResponse struct {
	Complete bool                           `json:"complete"`
	Summary  *model.PairwiseAnalysisSummary `json:"summary"`
	Results  []*model.PairwiseAnalysis      `json:"results"`
}

// Get the result of a pairwise analysis for the specified submissions.
func HandlePairwise(request *PairwiseRequest) (*PairwiseResponse, *core.APIError) {
	fullSubmissionIDs, courses, userErrors, systemErrors := analysis.ResolveSubmissionSpecs(request.SubmissionSpecs)

	if systemErrors != nil {
		return nil, core.NewUserContextInternalError("-619", &request.APIRequestUserContext, "Failed to resolve submission specs.").
			Err(systemErrors)
	}

	if userErrors != nil {
		return nil, core.NewBadUserRequestError("-620", &request.APIRequestUserContext,
			fmt.Sprintf("Failed to resolve submission specs: '%s'.", userErrors.Error())).
			Err(userErrors)
	}

	if !checkPermissions(request.ServerUser, courses) {
		return nil, core.NewBadUserRequestError("-621", &request.APIRequestUserContext,
			"User does not have permissions (server admin or course admin in all present courses.")
	}

	results, pendingCount, err := analysis.PairwiseAnalysis(fullSubmissionIDs, request.WaitForCompletion, request.ServerUser.Email)
	if err != nil {
		return nil, core.NewUserContextInternalError("-622", &request.APIRequestUserContext, "Failed to perform pairwise analysis.").
			Err(err)
	}

	response := PairwiseResponse{
		Complete: (pendingCount == 0),
		Summary:  model.NewPairwiseAnalysisSummary(results, pendingCount),
		Results:  results,
	}

	return &response, nil
}

func checkPermissions(user *model.ServerUser, courses []string) bool {
	// Admins can do whatever they want.
	if user.Role >= model.ServerRoleAdmin {
		return true
	}

	// Regular server users need to be at least an admin in every course they are making requests for.
	for _, course := range courses {
		if user.GetCourseRole(course) < model.CourseRoleAdmin {
			return false
		}
	}

	return true
}
