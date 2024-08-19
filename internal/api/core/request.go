package core

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// An api request that has been reflexively verifed.
// Once validated, callers should feel safe calling reflection methods on this without extra checks.
type ValidAPIRequest any

var RootUserNonces sync.Map

type APIRequest struct {
	// These are not provided in JSON, they are filled in during validation.
	RequestID string           `json:"-"`
	Endpoint  string           `json:"-"`
	Timestamp common.Timestamp `json:"-"`

	// This request is being used as part of a test.
	TestingMode bool `json:"-"`
}

// Context for a request that has a user (pretty much the lowest level of request).
type APIRequestUserContext struct {
	APIRequest

	UserEmail     string `json:"user-email"`
	UserPass      string `json:"user-pass"`
	RootUserNonce string `json:"root-user-nonce"`

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

	Assignment *model.Assignment
}

func (this *APIRequest) Validate(request any, endpoint string) *APIError {
	this.RequestID = util.UUID()
	this.Endpoint = endpoint
	this.Timestamp = common.NowTimestamp()

	this.TestingMode = config.TESTING_MODE.Get()

	return nil
}

// Validate that all the fields are populated correctly and
// that they are valid in the context of this server.
// Additionally, all context fields will be populated.
// This means that this request will be authenticated here.
// The full request (object that this is embedded in) is also sent.
func (this *APIRequestUserContext) Validate(request any, endpoint string) *APIError {
	apiErr := this.APIRequest.Validate(request, endpoint)
	if apiErr != nil {
		return apiErr
	}

	// Check for a valid nonce and skip auth if it exists.
	_, rootUserExists := RootUserNonces.LoadAndDelete(this.RootUserNonce)

	if this.RootUserNonce != "" {
		if !rootUserExists {
			return NewBadRequestError("-048", &this.APIRequest, "Incorrect root user nonce.")
		}

		this.ServerUser = &model.FakeRootUser
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
func (this *APIRequestCourseUserContext) Validate(request any, endpoint string) *APIError {
	apiErr := this.APIRequestUserContext.Validate(request, endpoint)
	if apiErr != nil {
		return apiErr
	}

	if this.CourseID == "" {
		return NewBadRequestError("-015", &this.APIRequest, "No course ID specified.")
	}

	var err error
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
func (this *APIRequestAssignmentContext) Validate(request any, endpoint string) *APIError {
	apiErr := this.APIRequestCourseUserContext.Validate(request, endpoint)
	if apiErr != nil {
		return apiErr
	}

	if this.AssignmentID == "" {
		return NewBadRequestError("-021", &this.APIRequest, "No assignment ID specified.")
	}

	id, err := common.ValidateID(this.AssignmentID)
	if err != nil {
		return NewBadRequestError("-035", &this.APIRequest, fmt.Sprintf("Provided assignment ID is invalid: '%s'.", this.AssignmentID)).
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

// Take in a pointer to an API request.
// Ensure this request has a type of known API request embedded in it and validate that embedded request.
func ValidateAPIRequest(request *http.Request, apiRequest any, endpoint string) *APIError {
	reflectPointer := reflect.ValueOf(apiRequest)
	if reflectPointer.Kind() != reflect.Pointer {
		return NewBareInternalError("-023", endpoint, "ValidateAPIRequest() must be called with a pointer.").
			Add("kind", reflectPointer.Kind().String())
	}

	// Ensure the request has an request type embedded, and validate it.
	foundRequestStruct, apiErr := validateRequestStruct(apiRequest, endpoint)
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

func validateRequestStruct(request any, endpoint string) (bool, *APIError) {
	// Check all the fields (including embedded ones) for structures that we recognize as requests.
	foundRequestStruct := false

	reflectValue := reflect.ValueOf(request).Elem()
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

			apiErr := apiRequest.Validate(request, endpoint)
			if apiErr != nil {
				return false, apiErr
			}

			fieldValue.Set(reflect.ValueOf(apiRequest))
		} else if fieldValue.Type() == reflect.TypeOf((*APIRequestUserContext)(nil)).Elem() {
			// APIRequestUserContext
			userRequest := fieldValue.Interface().(APIRequestUserContext)
			foundRequestStruct = true

			apiErr := userRequest.Validate(request, endpoint)
			if apiErr != nil {
				return false, apiErr
			}

			fieldValue.Set(reflect.ValueOf(userRequest))
		} else if fieldValue.Type() == reflect.TypeOf((*APIRequestCourseUserContext)(nil)).Elem() {
			// APIRequestCourseUserContext
			courseUserRequest := fieldValue.Interface().(APIRequestCourseUserContext)
			foundRequestStruct = true
			apiErr := courseUserRequest.Validate(request, endpoint)
			if apiErr != nil {
				return false, apiErr
			}

			fieldValue.Set(reflect.ValueOf(courseUserRequest))
		} else if fieldValue.Type() == reflect.TypeOf((*APIRequestAssignmentContext)(nil)).Elem() {
			// APIRequestAssignmentContext
			assignmentRequest := fieldValue.Interface().(APIRequestAssignmentContext)
			foundRequestStruct = true

			apiErr := assignmentRequest.Validate(request, endpoint)
			if apiErr != nil {
				return false, apiErr
			}

			fieldValue.Set(reflect.ValueOf(assignmentRequest))
		}
	}

	return foundRequestStruct, nil
}
