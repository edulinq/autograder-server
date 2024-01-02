package submission

import (
    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/api/core"
    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
)

type SubmitRequest struct {
    core.APIRequestAssignmentContext
    core.MinRoleStudent
    Files core.POSTFiles

    Message string `json:"message"`
}

type SubmitResponse struct {
    Rejected bool `json:"rejected"`
    Message string `json:"message"`

    GradingSucess bool `json:"grading-success"`
    GradingInfo *model.GradingInfo `json:"result"`
}

func HandleSubmit(request *SubmitRequest) (*SubmitResponse, *core.APIError) {
    response := SubmitResponse{};

    result, reject, err := grader.GradeDefault(request.Assignment, request.Files.TempDir, request.User.Email, request.Message);
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

    if (reject != nil) {
        response.Rejected = true;
        response.Message = reject.String();
        return &response, nil;
    }

    response.GradingSucess = true;
    response.GradingInfo = result.Info;

    return &response, nil;
}
