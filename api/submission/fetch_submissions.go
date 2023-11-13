package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/model"
)

type FetchSubmissionsRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleGrader

    FilterRole model.UserRole `json:"filter-role"`
}

type FetchSubmissionsResponse struct {
    GradingResults map[string]*model.GradingResult `json:"grading-results"`
}

func HandleFetchSubmissions(request *FetchSubmissionsRequest) (*FetchSubmissionsResponse, *core.APIError) {
    results, err := db.GetRecentSubmissionContents(request.Assignment, request.FilterRole);
    if (err != nil) {
        return nil, core.NewInternalError("-412", &request.APIRequestCourseUserContext, "Failed to get submissions.").
                Err(err);
    }

    return &FetchSubmissionsResponse{results}, nil;
}
