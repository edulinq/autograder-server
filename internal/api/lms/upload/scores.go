package upload

import (
	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/lms"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
)

type UploadScoresRequest struct {
	core.APIRequestCourseUserContext
	core.MinCourseRoleGrader
	Users core.CourseUsers `json:"-"`

	AssignmentLMSID core.NonEmptyString `json:"assignment-lms-id" required:""`
	Scores          []ScoreEntry        `json:"scores" required:""`
}

type ScoreEntry struct {
	Email string  `json:"email"`
	Score float64 `json:"score"`
}

type UploadScoresResponse struct {
	Count      int `json:"count"`
	ErrorCount int `json:"error-count"`

	UnrecognizedUsers []RowEntry `json:"unrecognized-users"`
	NoLMSIDUsers      []RowEntry `json:"no-lms-id-users"`
}

type RowEntry struct {
	Row   int `json:"row"`
	Entry any `json:"entry"`
}

// Upload scores from a tab-separated file to the course's LMS.
// The file should not have headers, and should have two columns: email and score.
func HandleUploadScores(request *UploadScoresRequest) (*UploadScoresResponse, *core.APIError) {
	if request.Course.GetLMSAdapter() == nil {
		return nil, core.NewBadRequestError("-405", request, "Course is not linked to an LMS.")
	}

	response := UploadScoresResponse{
		UnrecognizedUsers: []RowEntry{},
		NoLMSIDUsers:      []RowEntry{},
	}

	scores := parseScores(request, &response)

	if response.Count == 0 {
		return &response, nil
	}

	err := lms.UpdateAssignmentScores(request.Course, string(request.AssignmentLMSID), scores)
	if err != nil {
		return nil, core.NewInternalError("-406", request,
			"Failed to upload LMS scores.").Err(err)
	}

	return &response, nil
}

func parseScores(request *UploadScoresRequest, response *UploadScoresResponse) []*lmstypes.SubmissionScore {
	scores := make([]*lmstypes.SubmissionScore, 0, len(request.Scores))

	for i, entry := range request.Scores {
		user := request.Users[entry.Email]
		if user == nil {
			response.UnrecognizedUsers = append(response.UnrecognizedUsers, RowEntry{i, entry.Email})
			continue
		}

		if user.GetLMSID() == "" {
			response.NoLMSIDUsers = append(response.NoLMSIDUsers, RowEntry{i, entry.Email})
			continue
		}

		scores = append(scores, &lmstypes.SubmissionScore{
			UserID: user.GetLMSID(),
			Score:  entry.Score,
		})
	}

	response.Count = len(scores)
	response.ErrorCount = len(response.UnrecognizedUsers) + len(response.NoLMSIDUsers)

	return scores
}
