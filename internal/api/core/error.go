package core

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
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
	Sender       string
	Timestamp    timestamp.Timestamp
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
	args := make([]any, 0)

	args = append(args,
		this.SourceError,
		log.NewAttr("api-request-id", this.RequestID),
		log.NewAttr("locator", this.Locator),
		log.NewAttr("api-endpoint", this.Endpoint),
		log.NewAttr("sender", this.Sender),
		log.NewAttr("timestamp", this.Timestamp),
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

	version, err := util.GetFullCachedVersion()
	if err != nil {
		log.Warn("Failed to get the autograder version.", err)
	}

	return &APIResponse{
		ID:             this.RequestID,
		Locator:        locator,
		ServerVersion:  version,
		StartTimestamp: this.Timestamp,
		EndTimestamp:   timestamp.Now(),
		HTTPStatus:     this.HTTPStatus,
		Success:        (this.HTTPStatus == HTTP_STATUS_GOOD),
		Message:        this.ResponseText,
		Content:        nil,
	}
}

// Constructors for common cases.

func NewAuthError(locator string, requestContext any, internalMessage string) *APIError {
	apiError := &APIError{
		Locator:      locator,
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_STATUS_AUTH_ERROR,
		InternalText: fmt.Sprintf("Authentication failure: '%s'.", internalMessage),
		ResponseText: "Authentication failure, check email and password.",
	}

	applyContext(apiError, requestContext)

	return apiError
}

func NewBadRequestError(locator string, requestContext any, message string) *APIError {
	apiError := &APIError{
		Locator:      locator,
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_STATUS_BAD_REQUEST,
		InternalText: message,
		ResponseText: message,
	}

	applyContext(apiError, requestContext)

	return apiError
}

func NewPermissionsError(locator string, requestContext any, minRole any, actualRole any, internalMessage string) *APIError {
	apiError := &APIError{
		Locator:      locator,
		LogLevel:     log.LevelInfo,
		HTTPStatus:   HTTP_PERMISSIONS_ERROR,
		InternalText: fmt.Sprintf("Insufficient Permissions: '%s'.", internalMessage),
		ResponseText: "You have insufficient permissions for the requested operation.",
	}

	applyContext(apiError, requestContext)

	apiError.Add("min-role", minRole)
	apiError.Add("actual-role", actualRole)

	return apiError
}

func NewInternalError(locator string, requestContext any, internalMessage string) *APIError {
	apiError := &APIError{
		Locator:      locator,
		LogLevel:     log.LevelError,
		HTTPStatus:   HTTP_STATUS_SERVER_ERROR,
		InternalText: internalMessage,
		ResponseText: "The server failed to process your request.",
	}

	applyContext(apiError, requestContext)

	if apiError.RequestID != "" {
		apiError.ResponseText = fmt.Sprintf("The server failed to process your request. Please contact an administrator with this ID '%s'.", apiError.RequestID)
	}

	return apiError
}

// Figure out the type of the passed in request context and appy it to the given error.
func applyContext(apiError *APIError, requestContext any) {
	if requestContext == nil {
		log.Error("APIError has a nil request context.", log.NewAttr("locator", apiError.Locator))
		return
	}

	// When the request context is a string, it must be an endpoint.
	stringContext, ok := requestContext.(string)
	if ok {
		apiError.Endpoint = stringContext
		apiError.Timestamp = timestamp.Now()
		return
	}

	// Now check for APIRequest embedded types.

	reflectValue := reflect.ValueOf(requestContext)
	if reflectValue.Kind() == reflect.Pointer {
		reflectValue = reflectValue.Elem()
	}

	if reflectValue.Kind() != reflect.Struct {
		log.Error("APIError request context is not a struct.", log.NewAttr("context", requestContext), log.NewAttr("locator", apiError.Locator))
		return
	}

	baseContextValue := reflectValue.FieldByName("APIRequest")
	if baseContextValue.IsValid() {
		baseContext := baseContextValue.Interface().(APIRequest)

		apiError.RequestID = baseContext.RequestID
		apiError.Endpoint = baseContext.Endpoint
		apiError.Sender = baseContext.Sender
		apiError.Timestamp = baseContext.Timestamp
	}

	userContextValue := reflectValue.FieldByName("APIRequestUserContext")
	if userContextValue.IsValid() {
		userContext := userContextValue.Interface().(APIRequestUserContext)

		apiError.UserEmail = userContext.UserEmail
	}

	courseContextValue := reflectValue.FieldByName("APIRequestCourseUserContext")
	if courseContextValue.IsValid() {
		courseContext := courseContextValue.Interface().(APIRequestCourseUserContext)

		apiError.CourseID = courseContext.CourseID
	}

	assignmentContextValue := reflectValue.FieldByName("APIRequestAssignmentContext")
	if assignmentContextValue.IsValid() {
		assignmentContext := assignmentContextValue.Interface().(APIRequestAssignmentContext)

		apiError.AssignmentID = assignmentContext.AssignmentID
	}
}
