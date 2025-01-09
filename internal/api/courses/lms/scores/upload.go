package scores

import (
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/scoring"
)

type UploadRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleAdmin

	DryRun bool `json:"dry-run"`
}

type UploadResponse struct {
	DryRun  bool                         `json:"dry-run"`
	Results []*model.ExternalScoringInfo `json:"results"`
}

// Perform a full scoring and upload scores to the course's LMS.
func HandleUpload(request *UploadRequest) (*UploadResponse, *core.APIError) {
	scores, err := scoring.FullCourseScoringAndUpload(request.Course, request.DryRun)
	if err != nil {
		return nil, core.NewInternalError("-617", &request.APIRequestCourseUserContext,
			"Failed to perform a full course scoring.").Err(err)
	}

	results := make([]*model.ExternalScoringInfo, 0)
	for assignmentID, scoringInfos := range scores {
		for email, scoringInfo := range scoringInfos {
			results = append(results, scoringInfo.ToExternal(email, assignmentID))
		}
	}

	slices.SortFunc(results, func(a *model.ExternalScoringInfo, b *model.ExternalScoringInfo) int {
		result := strings.Compare(a.AssignmentID, b.AssignmentID)
		if result != 0 {
			return result
		}

		return strings.Compare(a.UserEmail, b.UserEmail)
	})

	response := &UploadResponse{
		DryRun:  request.DryRun,
		Results: results,
	}

	return response, nil
}
