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

    TargetEmail string `json:"target-email"`
    TargetSubmission string `json:"target-submission"`
}

type PeekResponse struct {
    Found bool `json:"found"`
    Assignment *artifact.GradedAssignment `json:"assignment"`
}

func HandlePeek(request *PeekRequest) (*PeekResponse, *core.APIError) {
    // Default the target to self.
    if (request.TargetEmail == "") {
        request.TargetEmail = request.User.Email;
    }

    // If the target is not self, user must be a grader or above.
    if ((request.TargetEmail != request.User.Email) && (request.User.Role < NON_SELF_PEEK_PERMISSIONS)) {
        return nil, core.NewBadPermissionsError("-401", &request.APIRequestCourseUserContext, NON_SELF_PEEK_PERMISSIONS, "Non-Self Peek");
    }

    // Ensure the submission ID is short.
    request.TargetSubmission = common.GetShortSubmissionID(request.TargetSubmission);

    response := PeekResponse{};

    paths, err := request.Assignment.GetSubmissionResults(request.TargetEmail);
    if (err != nil) {
        return nil, core.NewInternalError("-402", &request.APIRequestCourseUserContext, "Failed to get submission results.").Err(err);
    }

    if (len(paths) == 0) {
        return &response, nil;
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

    assignment := artifact.GradedAssignment{};
    err = util.JSONFromFile(targetPath, &assignment);
    if (err != nil) {
        return nil, core.NewInternalError("-403", &request.APIRequestCourseUserContext,
                "Failed to deserialize assignment.").Err(err).
                Add("path", targetPath);
    }

    response.Found = true;
    response.Assignment = &assignment;

    return &response, nil;
}
