package core

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

// An api request that has been reflexively verifed.
// Once validated, callers should feel safe calling reflection methods on this without extra checks.
type ValidAPIRequest any

// A random nonce is generated for each root user request (e.g. CMDs).
// The nonce is stored in RootUserNonces and is attached to the request.
// It's later validated when processing the request through the http socket and then immediately deleted
// to confirm the request came from a valid root user through the unix socket.
var RootUserNonces sync.Map

type APIRequest struct {
	// These are not provided in JSON, they are filled in during validation.
	RequestID string              `json:"-"`
	Endpoint  string              `json:"-"`
	Sender    string              `json:"-"`
	Timestamp timestamp.Timestamp `json:"-"`
	Context   context.Context     `json:"-"`
}

// Context for a request that has a user (pretty much the lowest level of request).
type APIRequestUserContext struct {
	APIRequest

	UserEmail     string `json:"user-email"`
	UserPass      string `json:"user-pass"`
	RootUserNonce string `json:"root-user-nonce,omitempty"`

	ServerUser *model.ServerUser `json:"-"`
}

// Context for a request that has a course and user from that course.
type APIRequestCourseUserContext struct {
	APIRequestUserContext

	CourseID string `json:"course-id"`

	Course *model.Course     `json:"-"`
	User   *model.CourseUser `json:"-"`
}

// Context for requests that need an assignment on top of a user/course.
type APIRequestAssignmentContext struct {
	APIRequestCourseUserContext

	AssignmentID string `json:"assignment-id"`

	Assignment *model.Assignment `json:"-"`
}

func (this *APIRequest) Validate(httpRequest *http.Request, request any, endpoint string) *APIError {
	this.RequestID = util.UUID()
	this.Endpoint = endpoint
	this.Timestamp = timestamp.Now()

	if httpRequest == nil {
		this.Context = context.Background()
	} else {
		this.Context = httpRequest.Context()
		this.Sender = httpRequest.RemoteAddr
	}

	return nil
}

// Validate that all the fields are populated correctly and
// that they are valid in the context of this server.
// Additionally, all context fields will be populated.
// This means that this request will be authenticated here.
// The full request (object that this is embedded in) is also sent.
func (this *APIRequestUserContext) Validate(httpRequest *http.Request, request any, endpoint string) *APIError {
	apiErr := this.APIRequest.Validate(httpRequest, request, endpoint)
	if apiErr != nil {
		return apiErr
	}

	if this.RootUserNonce != "" {
		// Check for a valid nonce and skip auth if it exists.
		_, rootUserExists := RootUserNonces.LoadAndDelete(this.RootUserNonce)
		if !rootUserExists {
			return NewAuthBadRequestError("-048", this, "Incorrect root user nonce.")
		}

		rootUser, err := db.GetServerUser(model.RootUserEmail)
		if err != nil {
			return NewUserContextInternalError("-049", this, "Failed to get the root user.")
		}

		if rootUser == nil {
			return NewUserContextInternalError("-050", this, "Root user not found.")
		}

		this.UserEmail = rootUser.Email
		this.ServerUser = rootUser
	} else {
		if this.UserEmail == "" {
			return NewBadRequestError("-016", &this.APIRequest, "No user email specified.")
		}

		if this.UserPass == "" {
			return NewBadRequestError("-017", &this.APIRequest, "No user password specified.")
		}

		this.ServerUser, apiErr = this.Auth()
		if apiErr != nil {
			return apiErr
		}
	}

	minRole, foundRole := getMaxServerRole(request)
	if !foundRole {
		minRole = model.ServerRoleUser
	}

	if this.ServerUser.Role < minRole {
		return NewBadServerPermissionsError("-041", this, minRole, "Base API Request")
	}

	return nil
}

// See APIRequestUserContext.Validate().
// The server user will be converted into a course user to be stored within this request.
func (this *APIRequestCourseUserContext) Validate(httpRequest *http.Request, request any, endpoint string) *APIError {
	apiErr := this.APIRequestUserContext.Validate(httpRequest, request, endpoint)
	if apiErr != nil {
		return apiErr
	}

	if this.CourseID == "" {
		return NewBadRequestError("-015", &this.APIRequest, "No course ID specified.")
	}

	id, err := common.ValidateID(this.CourseID)
	if err != nil {
		return NewBadRequestError("-052", &this.APIRequest,
			fmt.Sprintf("Could not find course (course ID ('%s') is invalid).", this.CourseID)).
			Course(this.CourseID).Err(err)
	}

	this.CourseID = id

	this.Course, err = db.GetCourse(this.CourseID)
	if err != nil {
		return NewInternalError("-032", this, "Unable to get course").Err(err)
	}

	if this.Course == nil {
		return NewBadRequestError("-018", &this.APIRequest, fmt.Sprintf("Could not find course: '%s'.", this.CourseID)).
			Course(this.CourseID)
	}

	this.User, err = this.ServerUser.ToCourseUser(this.Course.ID, true)
	if err != nil {
		return NewInternalError("-039", this, "Unable to convert server user to course user.").Err(err)
	}

	if this.User == nil {
		return NewBadRequestError("-040", &this.APIRequest, fmt.Sprintf("User '%s' is not enolled in course '%s'.", this.UserEmail, this.CourseID))
	}

	minRole, foundRole := getMaxCourseRole(request)
	if !foundRole {
		return NewInternalError("-019", this, "No role found for request. All course-based request structs require a minimum role.")
	}

	if this.User.Role < minRole {
		return NewBadCoursePermissionsError("-020", this, minRole, "Base API Request")
	}

	return nil
}

// See APIRequestUserContext.Validate().
func (this *APIRequestAssignmentContext) Validate(httpRequest *http.Request, request any, endpoint string) *APIError {
	apiErr := this.APIRequestCourseUserContext.Validate(httpRequest, request, endpoint)
	if apiErr != nil {
		return apiErr
	}

	if this.AssignmentID == "" {
		return NewBadRequestError("-021", &this.APIRequest, "No assignment ID specified.")
	}

	id, err := common.ValidateID(this.AssignmentID)
	if err != nil {
		return NewBadRequestError("-035", &this.APIRequest,
			fmt.Sprintf("Could not find assignment (assignment ID ('%s') is invalid).", this.AssignmentID)).
			Course(this.CourseID).Assignment(this.AssignmentID).Err(err)
	}

	this.AssignmentID = id

	this.Assignment = this.Course.GetAssignment(this.AssignmentID)
	if this.Assignment == nil {
		return NewBadRequestError("-022", &this.APIRequest, fmt.Sprintf("Could not find assignment: '%s'.", this.AssignmentID)).
			Course(this.CourseID).Assignment(this.AssignmentID)
	}

	return nil
}

// Convert an API request into loggable attributes.
// Note that this cannot be done (easily) with LogValue()
// since we would need to attach that method to each instead of an APIRequest.
func getLogAttributesFromAPIRequest(apiRequest ValidAPIRequest) []any {
	typedAPIRequest := reflect.ValueOf(apiRequest).Elem().FieldByName("APIRequest").Interface().(APIRequest)

	generalData, err := util.ToJSONMap(apiRequest)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to convert APIRequest to JSON: '%#v'.", apiRequest), err)

		return []any{
			log.NewAttr("id", typedAPIRequest.RequestID),
			log.NewAttr("endpoint", typedAPIRequest.Endpoint),
			log.NewAttr("sender", typedAPIRequest.Sender),
			log.NewAttr("timestamp", typedAPIRequest.Timestamp),
			log.NewAttr("conversion-error", err.Error()),
		}
	}

	// Clean up the data.

	// Add fields that do not have JSON equivalents.
	generalData["id"] = typedAPIRequest.RequestID
	generalData["endpoint"] = typedAPIRequest.Endpoint
	generalData["sender"] = typedAPIRequest.Sender

	// Remove passwords.
	delete(generalData, "user-pass")
	delete(generalData, "new-pass")
	cleanRawUsers(generalData)

	// Swap over standard loggable keys.
	// Note that these will be nil if they don't exist (and will be removed later).
	generalData[log.KEY_USER] = generalData["user-email"]
	delete(generalData, "user-email")
	generalData[log.KEY_COURSE] = generalData["course-id"]
	delete(generalData, "course-id")
	generalData[log.KEY_ASSIGNMENT] = generalData["assignment-id"]
	delete(generalData, "assignment-id")

	for key, value := range generalData {
		// Remove roles.
		if minRoleRegex.MatchString(key) {
			delete(generalData, key)
			continue
		}

		// Remove nils.
		if value == nil {
			delete(generalData, key)
			continue
		}
	}

	// Convert to loggable attributes.
	result := make([]any, 0, len(generalData))
	for key, value := range generalData {
		result = append(result, log.NewAttr(key, value))
	}

	return result
}

// Take in a pointer to an API request.
// Ensure this request has a type of known API request embedded in it and validate that embedded request.
func ValidateAPIRequest(request *http.Request, apiRequest any, endpoint string) *APIError {
	reflectPointer := reflect.ValueOf(apiRequest)
	if reflectPointer.Kind() != reflect.Pointer {
		return NewBareInternalError("-023", endpoint, "ValidateAPIRequest() must be called with a pointer.").
			Add("kind", reflectPointer.Kind().String())
	}

	// Ensure the request has an request type embedded, and validate it.
	foundRequestStruct, apiErr := validateRequestStruct(request, apiRequest, endpoint)
	if apiErr != nil {
		return apiErr
	}

	if !foundRequestStruct {
		return NewBareInternalError("-024", endpoint, "Request is not any kind of known API request.")
	}

	// Check for any special field types that we know how to populate.
	apiErr = checkRequestSpecialFields(request, apiRequest, endpoint)
	if apiErr != nil {
		return apiErr
	}

	return nil
}

// Cleanup any resources after the response has been sent.
// This function will return an error on failure,
// but the error will generally be ignored (since this will typically be called in a defer).
// So, any error will be logged here.
func CleanupAPIrequest(apiRequest ValidAPIRequest) error {
	reflectValue := reflect.ValueOf(apiRequest).Elem()

	for i := 0; i < reflectValue.NumField(); i++ {
		fieldValue := reflectValue.Field(i)

		if fieldValue.Type() == reflect.TypeOf((*POSTFiles)(nil)).Elem() {
			apiErr := cleanPostFiles(apiRequest, i)
			if apiErr != nil {
				return apiErr
			}
		}
	}

	return nil
}

func validateRequestStruct(httpRequest *http.Request, rawRequest any, endpoint string) (bool, *APIError) {
	// Check all the fields (including embedded ones) for structures that we recognize as requests.
	foundRequestStruct := false

	reflectValue := reflect.ValueOf(rawRequest).Elem()
	if reflectValue.Kind() != reflect.Struct {
		return false, NewBareInternalError("-031", endpoint, "Request's type must be a struct.").
			Add("kind", reflectValue.Kind().String())
	}

	for i := 0; i < reflectValue.NumField(); i++ {
		fieldValue := reflectValue.Field(i)

		if fieldValue.Type() == reflect.TypeOf((*APIRequest)(nil)).Elem() {
			// APIRequest
			apiRequest := fieldValue.Interface().(APIRequest)
			foundRequestStruct = true

			apiErr := apiRequest.Validate(httpRequest, rawRequest, endpoint)
			if apiErr != nil {
				return false, apiErr
			}

			fieldValue.Set(reflect.ValueOf(apiRequest))
		} else if fieldValue.Type() == reflect.TypeOf((*APIRequestUserContext)(nil)).Elem() {
			// APIRequestUserContext
			userRequest := fieldValue.Interface().(APIRequestUserContext)
			foundRequestStruct = true

			apiErr := userRequest.Validate(httpRequest, rawRequest, endpoint)
			if apiErr != nil {
				return false, apiErr
			}

			fieldValue.Set(reflect.ValueOf(userRequest))
		} else if fieldValue.Type() == reflect.TypeOf((*APIRequestCourseUserContext)(nil)).Elem() {
			// APIRequestCourseUserContext
			courseUserRequest := fieldValue.Interface().(APIRequestCourseUserContext)
			foundRequestStruct = true

			apiErr := courseUserRequest.Validate(httpRequest, rawRequest, endpoint)
			if apiErr != nil {
				return false, apiErr
			}

			fieldValue.Set(reflect.ValueOf(courseUserRequest))
		} else if fieldValue.Type() == reflect.TypeOf((*APIRequestAssignmentContext)(nil)).Elem() {
			// APIRequestAssignmentContext
			assignmentRequest := fieldValue.Interface().(APIRequestAssignmentContext)
			foundRequestStruct = true

			apiErr := assignmentRequest.Validate(httpRequest, rawRequest, endpoint)
			if apiErr != nil {
				return false, apiErr
			}

			fieldValue.Set(reflect.ValueOf(assignmentRequest))
		}
	}

	return foundRequestStruct, nil
}

// Reflexively get the request ID and timestamp from a request.
func getRequestIDAndTimestamp(request ValidAPIRequest) (string, timestamp.Timestamp) {
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

// Get the endpoint, sender, userEmail, courseID, assignmentID, and locator
// from a ValidAPIRequest and an APIError, both of which may be nil.
func getRequestInfo(request ValidAPIRequest, apiError *APIError) (string, string, string, string, string, string) {
	endpoint, sender, userEmail, courseID, assignmentID := getBasicAPIRequestInfo(request)
	locator := ""

	if apiError != nil {
		endpoint = util.GetStringWithDefault(endpoint, apiError.Endpoint)
		courseID = util.GetStringWithDefault(courseID, apiError.CourseID)
		assignmentID = util.GetStringWithDefault(assignmentID, apiError.AssignmentID)
		userEmail = util.GetStringWithDefault(userEmail, apiError.UserEmail)
		locator = apiError.Locator
	}

	return endpoint, sender, userEmail, courseID, assignmentID, locator
}

// Reflexively get the endpoint, userEmail, courseID, and assignmentID from a ValidAPIRequest.
func getBasicAPIRequestInfo(request ValidAPIRequest) (string, string, string, string, string) {
	endpoint := ""
	sender := ""
	userEmail := ""
	courseID := ""
	assignmentID := ""

	if request == nil {
		return endpoint, sender, userEmail, courseID, assignmentID
	}

	reflectValue := reflect.ValueOf(request).Elem()

	endpointValue := reflectValue.FieldByName("Endpoint")
	if endpointValue.IsValid() {
		endpoint = fmt.Sprintf("%s", endpointValue.Interface())
	}

	senderValue := reflectValue.FieldByName("Sender")
	if senderValue.IsValid() {
		sender = fmt.Sprintf("%s", senderValue.Interface())
	}

	userEmailValue := reflectValue.FieldByName("UserEmail")
	if userEmailValue.IsValid() {
		userEmail = fmt.Sprintf("%s", userEmailValue.Interface())
	}

	courseIDValue := reflectValue.FieldByName("CourseID")
	if courseIDValue.IsValid() {
		courseID = fmt.Sprintf("%s", courseIDValue.Interface())
	}

	assignmentIDValue := reflectValue.FieldByName("AssignmentID")
	if assignmentIDValue.IsValid() {
		assignmentID = fmt.Sprintf("%s", assignmentIDValue.Interface())
	}

	return endpoint, sender, userEmail, courseID, assignmentID
}

// Remove sensitive data from a generalized (from JSON) API request.
func cleanRawUsers(generalData map[string]any) {
	rawRawUsers, ok := generalData["raw-users"]
	if !ok {
		return
	}

	rawUsers, ok := rawRawUsers.([]any)
	if !ok {
		return
	}

	for _, rawRawUser := range rawUsers {
		rawUser, ok := rawRawUser.(map[string]any)
		if !ok {
			continue
		}

		delete(rawUser, "pass")
	}
}
