package api

import (
    "reflect"

    "github.com/eriq-augustine/autograder/util"
)

type APIResponse struct {
    ID string `json:"id"`
    StartTimestamp string `json:"start-timestamp"`
    EndTimestamp string `json:"end-timestamp"`

    HTTPStatus int `json:"status"`
    Success bool `json:"success"`

    Message string `json:"message"`
    Content any `json:"content"`
}

func NewAPIResponse(request ValidAPIRequest, content any) *APIResponse {
    id, timestamp := getRequestInfo(request);

    return &APIResponse{
        ID: id,
        StartTimestamp: timestamp,
        EndTimestamp: util.NowTimestamp(),
        HTTPStatus: HTTP_STATUS_GOOD,
        Success: true,
        Message: "",
        Content: content,
    };
}

// Reflexively get the request ID and timestamp from a request.
func getRequestInfo(request ValidAPIRequest) (string, string) {
    id := "";
    timestamp := "";

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
        timestamp = timestampValue.Interface().(string);
    }

    return id, timestamp;
}
