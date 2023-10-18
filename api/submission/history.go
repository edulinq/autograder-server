package submission

import (
    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

const NON_SELF_HISTORY_PERMISSIONS = usr.Grader;

type HistoryRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent
    Users core.CourseUsers

    TargetEmail string `json:"target-email"`
}

type HistoryResponse struct {
    Found bool `json:"found:`
    History []*artifact.SubmissionSummary `json:"history"`
}

func HandleHistory(request *HistoryRequest) (*HistoryResponse, *core.APIError) {
    // Default the target to self.
    if (request.TargetEmail == "") {
        request.TargetEmail = request.User.Email;
    }

    // If the target is not self, user must be a grader or above.
    if ((request.TargetEmail != request.User.Email) && (request.User.Role < NON_SELF_HISTORY_PERMISSIONS)) {
        return nil, core.NewBadPermissionsError("-406", &request.APIRequestCourseUserContext, NON_SELF_HISTORY_PERMISSIONS, "Non-Self History");
    }

    response := HistoryResponse{
        Found: false,
        History: make([]*artifact.SubmissionSummary, 0),
    };

    if (request.Users[request.TargetEmail] == nil) {
        // User not found.
        return &response, nil;
    }

    response.Found = true;

    paths, err := request.Assignment.GetSubmissionSummaries(request.TargetEmail);
    if (err != nil) {
        return nil, core.NewInternalError("-407", &request.APIRequestCourseUserContext, "Failed to get submission summaries.").Err(err);
    }

    for _, path := range paths {
        summary := artifact.SubmissionSummary{};
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return nil, core.NewInternalError("-408", &request.APIRequestCourseUserContext,
                    "Failed to deserialize submission summary.").Err(err).Add("path", path);
        }

        response.History = append(response.History, &summary);
    }

    return &response, nil;
}
