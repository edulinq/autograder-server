package core

import (
	"fmt"
	"reflect"

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
	id, startTime := getRequestInfoForAPIResponse(request)

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

// Reflexively get the request ID and timestamp from a request.
func getRequestInfoForAPIResponse(request ValidAPIRequest) (string, timestamp.Timestamp) {
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

// Get request values from a ValidAPIRequest and an APIError, both of which may be nil.
func getRequestInfoForStats(request ValidAPIRequest, apiError *APIError) (string, string, string, string, string) {
	endpoint, courseID, assignmentID, userEmail, locator := getAPIRequestInfoForStats(request)

	if apiError != nil {
		endpoint = util.GetStringWithDefault(endpoint, apiError.Endpoint)
		courseID = util.GetStringWithDefault(courseID, apiError.CourseID)
		assignmentID = util.GetStringWithDefault(assignmentID, apiError.AssignmentID)
		userEmail = util.GetStringWithDefault(userEmail, apiError.UserEmail)
		locator = apiError.Locator
	}

	return endpoint, userEmail, courseID, assignmentID, locator
}

func getAPIRequestInfoForStats(request ValidAPIRequest) (string, string, string, string, string) {
	if request == nil {
		return "", "", "", "", ""
	}

	endpoint := ""
	courseID := ""
	assignmentID := ""
	userEmail := ""
	locator := ""

	reflectValue := reflect.ValueOf(request).Elem()

	endpointValue := reflectValue.FieldByName("Endpoint")
	if endpointValue.IsValid() {
		endpoint = fmt.Sprintf("%s", endpointValue.Interface())
	}

	courseIDValue := reflectValue.FieldByName("CourseID")
	if courseIDValue.IsValid() {
		courseID = fmt.Sprintf("%s", courseIDValue.Interface())
	}

	assignmentIDValue := reflectValue.FieldByName("AssignmentID")
	if assignmentIDValue.IsValid() {
		assignmentID = fmt.Sprintf("%s", assignmentIDValue.Interface())
	}

	userEmailValue := reflectValue.FieldByName("UserEmail")
	if userEmailValue.IsValid() {
		userEmail = fmt.Sprintf("%s", userEmailValue.Interface())
	}

	return endpoint, userEmail, courseID, assignmentID, locator
}
