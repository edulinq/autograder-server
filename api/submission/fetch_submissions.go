package submission

import (
    "github.com/eriq-augustine/autograder/api/core"

    /*
    "path/filepath"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/util"
    */
)

type FetchSubmissionsRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleGrader
    Users core.CourseUsers `json:"-"`
}

type FetchSubmissionsResponse struct {
    SubmissionIDs map[string]string `json:"submission-ids"`
    Contents string `json:"contents"`
}

func HandleFetchSubmissions(request *FetchSubmissionsRequest) (*FetchSubmissionsResponse, *core.APIError) {
    // TEST
    return nil, nil;

    /* TEST - This needs some work. We already have the contents in the DB, we just need to strip the input/ path/dir from it.

    response := FetchSubmissionsResponse{
        SubmissionIDs: make(map[string]string, len(request.Users)),
    };

    paths, err := request.Assignment.GetAllRecentSubmissionSummaries(request.Users);
    if (err != nil) {
        return nil, core.NewInternalError("-412", &request.APIRequestCourseUserContext, "Failed to get submissions.").
                Err(err);
    }

    // When testing, create deterministic zip files.
    deterministicZip := config.TESTING_MODE.Get();

    zipOperation := util.NewOngoingZipOperation(deterministicZip);
    defer zipOperation.Close();

    for email, path := range paths {
        if (path == "") {
            continue;
        }

        summary := artifact.SubmissionSummary{};
        err = util.JSONFromFile(path, &summary);
        if (err != nil) {
            return nil, core.NewInternalError("-413", &request.APIRequestCourseUserContext,
                    "Failed to deserialize submission summary.").Err(err).
                    Add("path", path);
        }

        submissionDir := filepath.Dir(filepath.Dir(path));
        submissionID := common.GetShortSubmissionID(summary.ID);

        err = zipOperation.AddDir(submissionDir, filepath.Join("submissions", email));
        if (err != nil) {
            return nil, core.NewInternalError("-414", &request.APIRequestCourseUserContext,
                    "Failed to add submission dir to zip.").
                    Err(err).Add("submission-dir", submissionDir);
        }

        response.SubmissionIDs[email] = submissionID;
    }

    response.Contents = util.Base64Encode(zipOperation.GetBytes());

    return &response, nil;
    */
}
