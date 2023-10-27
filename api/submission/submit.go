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
    RawOutput string `json:"raw-output"`
    Summary *artifact.SubmissionSummary `json:"summary"`
    SubmissionResult *artifact.GradedAssignment `json:"result"`
}

func HandleSubmit(request *SubmitRequest) (*SubmitResponse, *core.APIError) {
    response := SubmitResponse{};

    result, summary, output, err := grader.GradeDefault(request.Assignment, request.Files.TempDir, request.User.Email, request.Message);
    if (err != nil) {
        if (output != "") {
            log.Debug().Err(err).Str("output", output).Msg("Submission grading failed, but output exists.");
            response.RawOutput = output;
        }

        return &response, nil;
    }

    response.GradingSucess = true;
    response.Summary = summary;
    response.SubmissionResult = result;

    return &response, nil;
}
