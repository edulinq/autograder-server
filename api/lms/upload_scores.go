package lms

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/lms/lmstypes"
)

type UploadScoresRequest struct {
    core.APIRequestCourseUserContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`

    AssignmentLMSID core.NonEmptyString `json:"assignment-lms-id"`
    Scores []ScoreEntry `json:"scores"`
}

type ScoreEntry struct {
    Email string `json:"email"`
    Score float64 `json:"score"`
}

type UploadScoresResponse struct {
    Count int `json:"count"`
    ErrorCount int `json:"error-count"`

    UnrecognizedUsers []RowEntry `json:"unrecognized-users"`
    NoLMSIDUsers []RowEntry `json:"no-lms-id-users"`
}

type RowEntry struct {
    Row int `json:"row"`
    Entry any `json:"entry"`
}

func HandleUploadScores(request *UploadScoresRequest) (*UploadScoresResponse, *core.APIError) {
    if (request.Course.GetLMSAdapter() == nil) {
        return nil, core.NewBadRequestError("-405", &request.APIRequest, "Course is not linked to an LMS.").
                Add("course", request.Course.GetID());
    }

    response := UploadScoresResponse{};
    scores := parseScores(request, &response);

    if (response.Count == 0) {
        return &response, nil;
    }

    err := lms.UpdateAssignmentScores(request.Course, string(request.AssignmentLMSID), scores);
    if (err != nil) {
        return nil, core.NewInternalError("-406", &request.APIRequestCourseUserContext,
                "Failed to upload LMS scores.").Err(err);
    }

    return &response, nil;
}

func parseScores(request *UploadScoresRequest, response *UploadScoresResponse) []*lmstypes.SubmissionScore {
    scores := make([]*lmstypes.SubmissionScore, 0, len(request.Scores));

    for i, entry := range request.Scores {
        user := request.Users[entry.Email];
        if (user == nil) {
            response.UnrecognizedUsers = append(response.UnrecognizedUsers, RowEntry{i, entry.Email});
            continue;
        }

        if (user.LMSID == "") {
            response.NoLMSIDUsers = append(response.NoLMSIDUsers, RowEntry{i, entry.Email});
            continue;
        }

        scores = append(scores, &lmstypes.SubmissionScore{
            UserID: user.LMSID,
            Score: entry.Score,
        });
    }

    response.Count = len(scores);
    response.ErrorCount = len(response.UnrecognizedUsers) + len(response.NoLMSIDUsers);

    return scores;
}
