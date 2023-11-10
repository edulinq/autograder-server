package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/usr"
)

type FetchScoresRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`

    // Filter results to only users with this role.
    FilterRole usr.UserRole `json:"filter-role"`
}

type FetchScoresResponse struct {
    SubmissionInfos map[string]*artifact.SubmissionHistoryItem `json:"submission-infos"`
}

func HandleFetchScores(request *FetchScoresRequest) (*FetchScoresResponse, *core.APIError) {
    submissionInfos, err := db.GetRecentSubmissionSurvey(request.Assignment, request.FilterRole);
    if (err != nil) {
        return nil, core.NewInternalError("-404", &request.APIRequestCourseUserContext, "Failed to get submission summaries.").Err(err);
    }

    return &FetchScoresResponse{submissionInfos}, nil;
}
