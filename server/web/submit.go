package web

import (
    "net/http"
    "os"

    "github.com/eriq-augustine/autograder/model"
)

type SubmissionRequest struct {
    model.BaseAPIRequest
    Assignment string `json:assignment`
    dir string
}

func NewSubmissionRequest(request *http.Request) (*SubmissionRequest, error) {
    var apiRequest SubmissionRequest;

    err := model.APIRequestFromHTTP(&apiRequest, request);
    if (err != nil) {
        return nil, err;
    }

    // TODO(eriq): Auth here.

    apiRequest.dir, err = model.StoreRequestFiles(request);
    if (err != nil) {
        return nil, err;
    }

    return &apiRequest, nil;
}

func (this *SubmissionRequest) Close() error {
    return os.RemoveAll(this.dir);
}

func handleSubmit(submission *SubmissionRequest) (int, any, error) {
    // TEST
    return 0, "Did it!", nil;
}
