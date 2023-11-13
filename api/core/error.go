package core

import (
    "errors"
    "fmt"
    "net/http"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/common"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

// Constants for http status codes.
// These should be used instead of choosing codes directly, so we remain consistent.
const (
    // Everything went well and there were no errors.
    // Note that there is a difference between a failed request and a negative result from a request.
    HTTP_STATUS_GOOD = http.StatusOK;
    // The user sent a request that is incorrect in some way.
    // These will usually not make it to the handler.
    HTTP_STATUS_BAD_REQUEST = http.StatusBadRequest;
    // There was some error beyond the user's control.
    // This is out fault.
    // These will generally happen at the handler level
    // (since before that would be bad requests).
    HTTP_STATUS_SERVER_ERROR = http.StatusInternalServerError;
    // Some form of authentication error occurred.
    // Intentionally vague.
    HTTP_STATUS_AUTH_ERROR = http.StatusUnauthorized;
    // The users role is not high enough for the specific operation.
    // Can happen at the validation or handling phases.
    HTTP_PERMISSIONS_ERROR = http.StatusForbidden;
)

// This is technically an error,
// but should generally be treated as an APIError and not a general error.
type APIError struct {
    RequestID string
    Locator string
    Endpoint string
    Timestamp common.Timestamp
    HTTPStatus int
    InternalText string
    ResponseText string
    SourceError error

    AdditionalDetails map[string]any
}

func (this *APIError) Error() string {
    return fmt.Sprintf("API Error -- %s", util.BaseString(this));
}

func (this *APIError) Log() {
    log.Error().
            Str("api-request-id", this.RequestID).Str("locator", this.Locator).Str("api-endpoint", this.Endpoint).
            Str("timestamp", this.Timestamp.String()).
            Int("http-status", this.HTTPStatus).
            Err(this.SourceError).
            Str("internal-text", this.InternalText).Str("response-text", this.ResponseText).
            Any("additional-details", this.AdditionalDetails).
            Msg("API Error");
}

// Add additional context to this error.
func (this *APIError) Add(key string, value any) *APIError {
    if (this.AdditionalDetails == nil) {
        this.AdditionalDetails = make(map[string]any);
    }

    this.AdditionalDetails[key] = value;
    return this;
}

// Attatch/append a sourse error.
func (this *APIError) Err(err error) *APIError {
    if (this.SourceError == nil) {
        this.SourceError = err;
    } else {
        this.SourceError = errors.Join(this.SourceError, err);
    }

    return this;
}

func (this *APIError) ToResponse() *APIResponse {
    // Remove the locator for HTTP_STATUS_AUTH_ERROR.
    locator := this.Locator;
    if (this.HTTPStatus == HTTP_STATUS_AUTH_ERROR) {
        locator = "";
    }

    return &APIResponse{
        ID: this.RequestID,
        Locator: locator,
        ServerVersion: util.GetAutograderFullVersion(),
        StartTimestamp: this.Timestamp,
        EndTimestamp: common.NowTimestamp(),
        HTTPStatus: this.HTTPStatus,
        Success: (this.HTTPStatus == HTTP_STATUS_GOOD),
        Message: this.ResponseText,
        Content: nil,
    };
}

// Constructors for common cases.

func NewBadRequestError(locator string, request *APIRequest, message string) *APIError {
    return &APIError{
        RequestID: request.RequestID,
        Locator: locator,
        Endpoint: request.Endpoint,
        Timestamp: request.Timestamp,
        HTTPStatus: HTTP_STATUS_BAD_REQUEST,
        InternalText: message,
        ResponseText: message,
    };
}

// A bad request before the request was even parsed (usually a JSON error).
func NewBareBadRequestError(locator string, endpoint string, message string) *APIError {
    return &APIError{
        RequestID: locator,
        Locator: locator,
        Endpoint: endpoint,
        Timestamp: common.NowTimestamp(),
        HTTPStatus: HTTP_STATUS_BAD_REQUEST,
        InternalText: message,
        ResponseText: message,
    };
}

func NewAuthBadRequestError(locator string, request *APIRequestCourseUserContext, internalMessage string) *APIError {
    err := &APIError{
        RequestID: request.RequestID,
        Locator: locator,
        Endpoint: request.Endpoint,
        Timestamp: request.Timestamp,
        HTTPStatus: HTTP_STATUS_AUTH_ERROR,
        InternalText: fmt.Sprintf("Authentication failure: '%s'.", internalMessage),
        ResponseText: "Authentication failure, check course, email, and password.",
    };

    err.Add("course", request.CourseID);
    err.Add("email", request.UserEmail);

    return err;
}

func NewBadPermissionsError(locator string, request *APIRequestCourseUserContext, minRole model.UserRole, internalMessage string) *APIError {
    err := &APIError{
        RequestID: request.RequestID,
        Locator: locator,
        Endpoint: request.Endpoint,
        Timestamp: request.Timestamp,
        HTTPStatus: HTTP_PERMISSIONS_ERROR,
        InternalText: fmt.Sprintf("Insufficient Permissions: '%s'.", internalMessage),
        ResponseText: "You have insufficient permissions for the requested operation.",
    };

    err.Add("course", request.CourseID);
    err.Add("email", request.UserEmail);

    err.Add("actual-role", request.User.Role);
    err.Add("min-required-role", minRole);

    return err;
}

func NewInternalError(locator string, request *APIRequestCourseUserContext, internalMessage string) *APIError {
    err := &APIError{
        RequestID: request.RequestID,
        Locator: locator,
        Endpoint: request.Endpoint,
        Timestamp: request.Timestamp,
        HTTPStatus: HTTP_STATUS_SERVER_ERROR,
        InternalText: internalMessage,
        ResponseText: fmt.Sprintf("The server failed to process your request. Please contact an adimistrator with this ID '%s'.", request.RequestID),
    };

    err.Add("course", request.CourseID);
    err.Add("email", request.UserEmail);

    return err;
}

// Very rare errors can occur so early that there is not even a request id.
func NewBareInternalError(locator string, endpoint string, internalMessage string) *APIError {
    err := &APIError{
        RequestID: locator,
        Locator: locator,
        Endpoint: endpoint,
        Timestamp: common.NowTimestamp(),
        HTTPStatus: HTTP_STATUS_SERVER_ERROR,
        InternalText: internalMessage,
        ResponseText: fmt.Sprintf("The server failed to process your request. Please contact an adimistrator with this ID '%s'.", locator),
    };

    return err;
}
