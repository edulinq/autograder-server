package core

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// Constants for http status codes.
// These should be used instead of choosing codes directly, so we remain consistent.
const (
	// Everything went well and there were no errors.
	// Note that there is a difference between a failed request and a negative result from a request.
	HTTP_STATUS_GOOD = http.StatusOK
	// The user sent a request that is incorrect in some way.
	// These will usually not make it to the handler.
	HTTP_STATUS_BAD_REQUEST = http.StatusBadRequest
	// There was some error beyond the user's control.
	// This is out fault.
	// These will generally happen at the handler level
	// (since before that would be bad requests).
	HTTP_STATUS_SERVER_ERROR = http.StatusInternalServerError
	// Some form of authentication error occurred.
	// Intentionally vague.
	HTTP_STATUS_AUTH_ERROR = http.StatusUnauthorized
	// The users role is not high enough for the specific operation.
	// Can happen at the validation or handling phases.
	HTTP_PERMISSIONS_ERROR = http.StatusForbidden
)

// This is technically an error,
// but should generally be treated as an APIError and not a general error.
type APIError struct {
	RequestID    string
	Locator      string
	Endpoint     string
	Timestamp    common.Timestamp
	LogLevel     log.LogLevel
	HTTPStatus   int
	InternalText string
	ResponseText string
	SourceError  error

	CourseID     string
	AssignmentID string
	UserEmail    string

	AdditionalDetails map[string]any
}

func (this *APIError) Error() string {
	return fmt.Sprintf("API Error -- %s", util.BaseString(this))
}

func (this *APIError) Log() {
	args := make([]any, 0, 16)

	args = append(args,
		this.SourceError,
		log.NewAttr("api-request-id", this.RequestID),
		log.NewAttr("locator", this.Locator),
		log.NewAttr("api-endpoint", this.Endpoint),
		log.NewAttr("timestamp", this.Timestamp.String()),
		log.NewAttr("http-status", this.HTTPStatus),
		log.NewAttr("internal-text", this.InternalText),
		log.NewAttr("response-text", this.ResponseText))

	for key, value := range this.AdditionalDetails {
		args = append(args, log.NewAttr(key, value))
	}

	if this.CourseID != "" {
		args = append(args, log.NewCourseAttr(this.CourseID))
	}

	if this.AssignmentID != "" {
		args = append(args, log.NewAssignmentAttr(this.AssignmentID))
	}

	if this.UserEmail != "" {
		args = append(args, log.NewUserAttr(this.UserEmail))
	}

	log.LogToLevel(this.LogLevel, "API Error", args...)
}

// Add additional context to this error.
func (this *APIError) Add(key string, value any) *APIError {
	if this.AdditionalDetails == nil {
		this.AdditionalDetails = make(map[string]any)
	}

	this.AdditionalDetails[key] = value
	return this
}

// Attatch a course ID.
func (this *APIError) Course(id string) *APIError {
	this.CourseID = id
	return this
}

// Attatch an assignment ID.
func (this *APIError) Assignment(id string) *APIError {
	this.AssignmentID = id
	return this
}

// Attatch a user email.
func (this *APIError) User(email string) *APIError {
	this.UserEmail = email
	return this
}

// Attatch/append a sourse error.
func (this *APIError) Err(err error) *APIError {
	if this.SourceError == nil {
		this.SourceError = err
	} else {
		this.SourceError = errors.Join(this.SourceError, err)
	}

	return this
}

func (this *APIError) ToResponse() *APIResponse {
	// Remove the locator for HTTP_STATUS_AUTH_ERROR.
	locator := this.Locator
	if this.HTTPStatus == HTTP_STATUS_AUTH_ERROR {
		locator = ""
	}

	return &APIResponse{
		ID:             this.RequestID,
		Locator:        locator,
		ServerVersion:  util.GetAutograderFullVersion(),
		StartTimestamp: this.Timestamp,
		EndTimestamp:   common.NowTimestamp(),
		HTTPStatus:     this.HTTPStatus,
		Success:        (this.HTTPStatus == HTTP_STATUS_GOOD),
		Message:        this.ResponseText,
		Content:        nil,
	}
}

// Constructors for common cases.

func NewBadRequestError(locator string, request *APIRequest, message string) *APIError {
	return &APIError{
		RequestID:    request.RequestID,
		Locator:      locator,
		Endpoint:     request.Endpoint,
		Timestamp:    request.Timestamp,
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_STATUS_BAD_REQUEST,
		InternalText: message,
		ResponseText: message,
	}
}

func NewBadCourseRequestError(locator string, request *APIRequestCourseUserContext, message string) *APIError {
	err := &APIError{
		RequestID:    request.RequestID,
		Locator:      locator,
		Endpoint:     request.Endpoint,
		Timestamp:    request.Timestamp,
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_STATUS_BAD_REQUEST,
		InternalText: message,
		ResponseText: message,
		CourseID:     request.CourseID,
		UserEmail:    request.UserEmail,
	}

	return err
}

// A bad request before the request was even parsed (usually a JSON error).
func NewBareBadRequestError(locator string, endpoint string, message string) *APIError {
	return &APIError{
		RequestID:    locator,
		Locator:      locator,
		Endpoint:     endpoint,
		Timestamp:    common.NowTimestamp(),
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_STATUS_BAD_REQUEST,
		InternalText: message,
		ResponseText: message,
	}
}

func NewAuthBadRequestError(locator string, request *APIRequestUserContext, internalMessage string) *APIError {
	err := &APIError{
		RequestID:    request.RequestID,
		Locator:      locator,
		Endpoint:     request.Endpoint,
		Timestamp:    request.Timestamp,
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_STATUS_AUTH_ERROR,
		InternalText: fmt.Sprintf("Authentication failure: '%s'.", internalMessage),
		ResponseText: "Authentication failure, check email and password.",
		UserEmail:    request.UserEmail,
	}

	return err
}

func NewBadServerPermissionsError(locator string, request *APIRequestUserContext, minRole model.ServerUserRole, internalMessage string) *APIError {
	err := &APIError{
		RequestID:    request.RequestID,
		Locator:      locator,
		Endpoint:     request.Endpoint,
		Timestamp:    request.Timestamp,
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_PERMISSIONS_ERROR,
		InternalText: fmt.Sprintf("Insufficient Permissions: '%s'.", internalMessage),
		ResponseText: "You have insufficient permissions for the requested operation.",
		UserEmail:    request.UserEmail,
	}

	err.Add("actual-server-role", request.ServerUser.Role)
	err.Add("min-required-server-role", minRole)

	return err
}

func NewBadCoursePermissionsError(locator string, request *APIRequestCourseUserContext, minRole model.CourseUserRole, internalMessage string) *APIError {
	err := &APIError{
		RequestID:    request.RequestID,
		Locator:      locator,
		Endpoint:     request.Endpoint,
		Timestamp:    request.Timestamp,
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_PERMISSIONS_ERROR,
		InternalText: fmt.Sprintf("Insufficient Permissions: '%s'.", internalMessage),
		ResponseText: "You have insufficient permissions for the requested operation.",
		CourseID:     request.CourseID,
		UserEmail:    request.UserEmail,
	}

	err.Add("actual-course-role", request.User.Role)
	err.Add("min-required-course-role", minRole)

	return err
}

func NewInternalError(locator string, request *APIRequestCourseUserContext, internalMessage string) *APIError {
	err := &APIError{
		RequestID:    request.RequestID,
		Locator:      locator,
		Endpoint:     request.Endpoint,
		Timestamp:    request.Timestamp,
		LogLevel:     log.LevelError,
		HTTPStatus:   HTTP_STATUS_SERVER_ERROR,
		InternalText: internalMessage,
		ResponseText: fmt.Sprintf("The server failed to process your request. Please contact an administrator with this ID '%s'.", request.RequestID),
		CourseID:     request.CourseID,
		UserEmail:    request.UserEmail,
	}

	return err
}

func NewUsertContextInternalError(locator string, request *APIRequestUserContext, internalMessage string) *APIError {
	err := &APIError{
		RequestID:    request.RequestID,
		Locator:      locator,
		Endpoint:     request.Endpoint,
		Timestamp:    request.Timestamp,
		LogLevel:     log.LevelError,
		HTTPStatus:   HTTP_STATUS_SERVER_ERROR,
		InternalText: internalMessage,
		ResponseText: fmt.Sprintf("The server failed to process your request. Please contact an administrator with this ID '%s'.", request.RequestID),
		UserEmail:    request.UserEmail,
	}

	return err
}

func NewBaseInternalError(locator string, request *APIRequest, internalMessage string) *APIError {
	err := &APIError{
		RequestID:    request.RequestID,
		Locator:      locator,
		Endpoint:     request.Endpoint,
		Timestamp:    request.Timestamp,
		LogLevel:     log.LevelError,
		HTTPStatus:   HTTP_STATUS_SERVER_ERROR,
		InternalText: internalMessage,
		ResponseText: fmt.Sprintf("The server failed to process your request. Please contact an administrator with this ID '%s'.", request.RequestID),
	}

	return err
}

// Very rare errors can occur so early that there is not even a request id.
func NewBareInternalError(locator string, endpoint string, internalMessage string) *APIError {
	err := &APIError{
		RequestID:    locator,
		Locator:      locator,
		Endpoint:     endpoint,
		Timestamp:    common.NowTimestamp(),
		LogLevel:     log.LevelError,
		HTTPStatus:   HTTP_STATUS_SERVER_ERROR,
		InternalText: internalMessage,
		ResponseText: fmt.Sprintf("The server failed to process your request. Please contact an administrator with this ID '%s'.", locator),
	}

	return err
}
