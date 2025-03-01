package core

import (
	"github.com/edulinq/autograder/internal/log"
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
	id, startTime := getRequestIDAndTimestamp(request)

	version, err := util.GetAutograderVersion()
	if err != nil {
		log.Warn("Failed to get the autograder version.", err)
	}

	return &APIResponse{
		ID:             id,
		ServerVersion:  version,
		StartTimestamp: startTime,
		EndTimestamp:   timestamp.Now(),
		HTTPStatus:     HTTP_STATUS_GOOD,
		Success:        true,
		Message:        "",
		Content:        content,
	}
}
