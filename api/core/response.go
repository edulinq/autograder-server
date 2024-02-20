package core

import (
    "reflect"

    "github.com/edulinq/autograder/common"
    "github.com/edulinq/autograder/util"
)

type APIResponse struct {
    ID string `json:"id"`
    Locator string `json:"locator"`
    ServerVersion string `json:"server-version"`

    StartTimestamp common.Timestamp `json:"start-timestamp"`
    EndTimestamp common.Timestamp `json:"end-timestamp"`

    HTTPStatus int `json:"status"`
    Success bool `json:"success"`

    Message string `json:"message"`
    Content any `json:"content"`
}

func (this *APIResponse) String() string {
    return util.BaseString(this);
}

func NewAPIResponse(request ValidAPIRequest, content any) *APIResponse {
    id, timestamp := getRequestInfo(request);

    return &APIResponse{
        ID: id,
        ServerVersion: util.GetAutograderFullVersion(),
        StartTimestamp: timestamp,
        EndTimestamp: common.NowTimestamp(),
        HTTPStatus: HTTP_STATUS_GOOD,
        Success: true,
        Message: "",
        Content: content,
    };
}

// Reflexively get the request ID and timestamp from a request.
func getRequestInfo(request ValidAPIRequest) (string, common.Timestamp) {
    id := "";
    timestamp := common.NowTimestamp();

    if (request == nil) {
        return id, timestamp;
    }

    reflectValue := reflect.ValueOf(request).Elem();

    idValue := reflectValue.FieldByName("RequestID");
    if (idValue.IsValid()) {
        id = idValue.Interface().(string);
    }

    timestampValue := reflectValue.FieldByName("Timestamp");
    if (timestampValue.IsValid()) {
        timestamp = timestampValue.Interface().(common.Timestamp);
    }

    return id, timestamp;
}
