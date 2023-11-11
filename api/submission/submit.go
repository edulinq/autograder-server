package submission

import (
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/grader"
)

type SubmitRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent
    Files core.POSTFiles

    Message string `json:"message"`
}

type SubmitResponse struct {
    GradingSucess bool `json:"grading-success"`
    SubmissionResult *artifact.GradedAssignment `json:"result"`
}

func HandleSubmit(request *SubmitRequest) (*SubmitResponse, *core.APIError) {
    response := SubmitResponse{};

    result, err := grader.GradeDefault(request.Assignment, request.Files.TempDir, request.User.Email, request.Message);
    if (err != nil) {
        stdout := "";
        stderr := "";

        if ((result != nil) && (result.HasTextOutput())) {
            stdout = result.Stdout;
            stderr = result.Stderr;
        }

        log.Debug().Err(err).Str("stdout", stdout).Str("stderr", stderr).Msg("Submission grading failed.");

        return &response, nil;
    }

    response.GradingSucess = true;
    response.SubmissionResult = result.Result;

    return &response, nil;
}
