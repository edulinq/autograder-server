package submission

import (
    "path/filepath"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
)

type FetchSubmissionRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent

    TargetUser core.TargetUserSelfOrGrader `json:"target-email"`
    TargetSubmission string `json:"target-submission"`
}

type FetchSubmissionResponse struct {
    FoundUser bool `json:"found-user"`
    FoundSubmission bool `json:"found-submission"`
    SubmissionID string `json:"submission-id"`
    Contents string `json:"contents"`
}

func HandleFetchSubmission(request *FetchSubmissionRequest) (*FetchSubmissionResponse, *core.APIError) {
    // Ensure the submission ID is short.
    request.TargetSubmission = common.GetShortSubmissionID(request.TargetSubmission);

    response := FetchSubmissionResponse{};

    if (!request.TargetUser.Found) {
        return &response, nil;
    }

    response.FoundUser = true;

    paths, err := request.Assignment.GetSubmissionSummaries(request.TargetUser.Email);
    if (err != nil) {
        return nil, core.NewInternalError("-409", &request.APIRequestCourseUserContext, "Failed to get submission summary.").
                Err(err).Add("email", request.TargetUser.Email);
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

    summary := artifact.SubmissionSummary{};
    err = util.JSONFromFile(targetPath, &summary);
    if (err != nil) {
        return nil, core.NewInternalError("-410", &request.APIRequestCourseUserContext,
                "Failed to deserialize submission summary.").Err(err).
                Add("path", targetPath);
    }

    // When testing, create deterministic zip files.
    deterministicZip := config.TESTING_MODE.Get();

    submissionPath := filepath.Dir(filepath.Dir(targetPath));
    data, err := util.ZipToBytes(submissionPath, summary.ID, deterministicZip);
    if (err != nil) {
        return nil, core.NewInternalError("-411", &request.APIRequestCourseUserContext,
                "Failed to create a zip file in memory for submission.").
                Err(err).Add("path", submissionPath);
    }

    response.FoundSubmission = true;
    response.SubmissionID = summary.ID;
    response.Contents = util.Base64Encode(data);

    return &response, nil;
}
