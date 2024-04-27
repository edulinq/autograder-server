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
    LateAcknowledgment *bool `json:"late-acknowledgment"`
    Files core.POSTFiles

    Message string `json:"message"`
}

type SubmitResponse struct {
    Rejected bool `json:"rejected"`
    RequireLateAcknowledgment bool `json:"require-late-acknowledgment"`
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

        log.Info("Submission grading failed.", err, request.Assignment, log.NewAttr("stdout", stdout), log.NewAttr("stderr", stderr), request.User);

        return &response, nil;
    }

    if (reject != nil) {
        log.Debug("Submission rejected.", request.Assignment, log.NewAttr("reason", reject.String()), log.NewAttr("request", request), request.User);

        response.Rejected = true;
        response.Message = reject.String();

        if _, ok := reject.(*grader.RejectMissingLateAcknowledgment); ok {
            response.RequireLateAcknowledgment = true;
        }
        return &response, nil;
    }

    response.GradingSucess = true;
    response.GradingInfo = result.Info;

    return &response, nil;
}
