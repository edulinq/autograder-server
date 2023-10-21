package submission

import (
    "path/filepath"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

const NON_SELF_PEEK_PERMISSIONS = usr.Grader;

type PeekRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent

    TargetUser core.TargetUserSelfOrGrader `json:"target-email"`
    TargetSubmission string `json:"target-submission"`
}

type PeekResponse struct {
    FoundUser bool `json:"found-user"`
    FoundSubmission bool `json:"found-submission"`
    Submission *artifact.GradedAssignment `json:"submission"`
}

func HandlePeek(request *PeekRequest) (*PeekResponse, *core.APIError) {
    // Ensure the submission ID is short.
    request.TargetSubmission = common.GetShortSubmissionID(request.TargetSubmission);

    response := PeekResponse{};

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;

    paths, err := request.Assignment.GetSubmissionResults(request.TargetUser.Email);
    if (err != nil) {
        return nil, core.NewInternalError("-402", &request.APIRequestCourseUserContext, "Failed to get submission results.").Err(err);
    }

    targetPath := "";
    // Start with the most recent submission and go backwards.
    for i := (len(paths) - 1); i >= 0; i-- {
        path := paths[i];
        id := filepath.Base(filepath.Dir(filepath.Dir(path)));

        if ((request.TargetSubmission == "") || (request.TargetSubmission == id)) {
            targetPath = path;
            break;
        }
    }

    if (targetPath == "") {
        return &response, nil;
    }

    submission := artifact.GradedAssignment{};
    err = util.JSONFromFile(targetPath, &submission);
    if (err != nil) {
        return nil, core.NewInternalError("-403", &request.APIRequestCourseUserContext,
                "Failed to deserialize submission.").Err(err).
                Add("path", targetPath);
    }

    response.FoundSubmission = true;
    response.Submission = &submission;

    return &response, nil;
}
