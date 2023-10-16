package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

type FetchScoresRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`

    // Filter results to only users with this role.
    FilterRole usr.UserRole `json:"filter-role"`
}

type FetchScoresResponse struct {
    Summaries map[string]*artifact.SubmissionSummary `json:"scores"`
}

func HandleFetchScores(request *FetchScoresRequest) (*FetchScoresResponse, *core.APIError) {
    paths, err := request.Assignment.GetAllRecentSubmissionSummaries(request.Users);
    if (err != nil) {
        return nil, core.NewInternalError("-404", &request.APIRequestCourseUserContext, "Failed to get submission summaries.").Err(err);
    }

    summaries := make(map[string]*artifact.SubmissionSummary, len(paths));

    for email, user := range request.Users {
        if ((request.FilterRole != usr.Unknown) && (request.FilterRole != user.Role)) {
            continue;
        }

        path := paths[email];
        if (path == "") {
            summaries[email] = nil;
            continue;
        }

        summary := artifact.SubmissionSummary{};
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return nil, core.NewInternalError("-405", &request.APIRequestCourseUserContext,
                    "Failed to deserialize submission summary.").Err(err).Add("path", path);
        }

        summaries[email] = &summary;
    }

    return &FetchScoresResponse{summaries}, nil;
}
