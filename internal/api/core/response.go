package core

import (
	"reflect"

	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type APIResponse struct {
	ID            string       `json:"id"`
	Locator       string       `json:"locator"`
	ServerVersion util.Version `json:"server-version"`

	StartTimestamp timestamp.Timestamp `json:"start-timestamp"`
	EndTimestamp   timestamp.Timestamp `json:"end-timestamp"`

	HTTPStatus int  `json:"status"`
	Success    bool `json:"success"`

	Message string `json:"message"`
	Content any    `json:"content"`
}

func (this *APIResponse) String() string {
	return util.BaseString(this)
}

func NewAPIResponse(request ValidAPIRequest, content any) *APIResponse {
	id, startTime := getRequestInfo(request)

	return &APIResponse{
		ID:             id,
		ServerVersion:  util.GetAutograderFullVersion(),
		StartTimestamp: startTime,
		EndTimestamp:   timestamp.Now(),
		HTTPStatus:     HTTP_STATUS_GOOD,
		Success:        true,
		Message:        "",
		Content:        content,
	}
}

// Reflexively get the request ID and timestamp from a request.
func getRequestInfo(request ValidAPIRequest) (string, timestamp.Timestamp) {
	id := ""
	startTime := timestamp.Now()

	if request == nil {
		return id, startTime
	}

	reflectValue := reflect.ValueOf(request).Elem()

	idValue := reflectValue.FieldByName("RequestID")
	if idValue.IsValid() {
		id = idValue.Interface().(string)
	}

	timestampValue := reflectValue.FieldByName("Timestamp")
	if timestampValue.IsValid() {
		startTime = timestampValue.Interface().(timestamp.Timestamp)
	}

	return id, startTime
}
