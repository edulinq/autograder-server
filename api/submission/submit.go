package submission

import (
    "github.com/edulinq/autograder/api/core"
    "github.com/edulinq/autograder/grader"
    "github.com/edulinq/autograder/log"
    "github.com/edulinq/autograder/model"
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

        log.Info("Submission grading failed.", err, request.Assignment, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr));

        return &response, nil;
    }

    if (reject != nil) {
        log.Debug("Submission rejected.", request.Assignment, log.NewAttr("reason", reject.String()), log.NewAttr("request", request));

        response.Rejected = true;
        response.Message = reject.String();
        return &response, nil;
    }

    response.GradingSucess = true;
    response.GradingInfo = result.Info;

    return &response, nil;
}
