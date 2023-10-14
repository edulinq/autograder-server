package api

import (
    "reflect"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
)

const (
    // Post form key for request content.
    API_REQUEST_CONTENT_KEY = "content";
)

// The minimum user roles required encoded as a type so it can be embedded into a request struct.
type MinRoleOwner bool;
type MinRoleAdmin bool;
type MinRoleGrader bool;
type MinRoleStudent bool;
type MinRoleOther bool;

// A request having a field of this type indicates that the users for the course should be automatically fetched.
// The existence of this type in a struct also indicates that the request is at least a APIRequestCourseUserContext.
type CourseUsers map[string]*usr.User;

type APIRequest struct {
    // These are not provided in JSON, they are filled in during validation.
    RequestID string `json:"-"`
    Endpoint string `json:"-"`
    Timestamp string `json:"-"`
}

// Context for a request that has a course and user (pretty much the lowest level of request).
type APIRequestCourseUserContext struct {
    APIRequest

    CourseID string `json:"course-id"`
    UserEmail string `json:"user-email"`
    UserPass string `json:"user-pass"`

    // These fields are filled out as the request is parsed,
    // before being sent to the handler.
    course *model.Course
    user *usr.User
}

//Context for requests that need an assignment on top of a user/course.
type APIRequestAssignmentContext struct {
    APIRequestCourseUserContext

    AssignmentID string `json:"assignment-id"`

    assignment *model.Assignment
}

func (this *APIRequest) Validate(request any, endpoint string) *APIError {
    this.RequestID = util.UUID();
    this.Endpoint = endpoint;
    this.Timestamp = util.NowTimestamp();

    return nil;
}

// Validate that all the fields are populated correctly and
// that they are valid in the context of this server,
// Additionally, all context fields will be populated.
// This means that this request will be authenticated here.
// The full request (object that this is embedded in) is also sent.
func (this *APIRequestCourseUserContext) Validate(request any, endpoint string) *APIError {
    apiErr := this.APIRequest.Validate(request, endpoint);
    if (apiErr != nil) {
        return apiErr;
    }

    if (this.CourseID == "") {
        return NewBadRequestError(&this.APIRequest, "No course ID specified.");
    }

    if (this.UserEmail == "") {
        return NewBadRequestError(&this.APIRequest, "No user email specified.");
    }

    if (this.UserPass == "") {
        return NewBadRequestError(&this.APIRequest, "No user password specified.");
    }

    this.course = grader.GetCourse(this.CourseID);
    if (this.course == nil) {
        return NewBadRequestError(&this.APIRequest, "Could not find course.").Add("course-id", this.CourseID);
    }

    this.user, apiErr = this.Auth();
    if (apiErr != nil) {
        return apiErr;
    }

    minRole, foundRole := getMaxRole(request);
    if (!foundRole) {
        return NewInternalError(this, "No role found for request. All request structs require a minimum role.");
    }

    if (this.user.Role < minRole) {
        return NewBadPermissionsError(this, minRole, "");
    }

    return nil;
}

// See APIRequestCourseUserContext.Validate().
func (this *APIRequestAssignmentContext) Validate(request any, endpoint string) *APIError {
    apiErr := this.APIRequestCourseUserContext.Validate(request, endpoint);
    if (apiErr != nil) {
        return apiErr;
    }

    if (this.AssignmentID == "") {
        return NewBadRequestError(&this.APIRequest, "No assignment ID specified.");
    }

    this.assignment = this.course.Assignments[this.AssignmentID];
    if (this.assignment == nil) {
        return NewBadRequestError(&this.APIRequest, "Could not find assignment.").
            Add("course-id", this.CourseID).Add("assignment-id", this.AssignmentID);
    }

    return nil;
}

// Take in a pointer to an API request.
// Ensure this request has a type of known API request embedded in it and validate that embedded request.
func ValidateAPIRequest(request any, endpoint string) *APIError {
    reflectPointer := reflect.ValueOf(request);
    if (reflectPointer.Kind() != reflect.Pointer) {
        return NewBareInternalError("-512", endpoint, "ValidateAPIRequest() must be called with a pointer.");
    }

    // Ensure the request has an request type embedded, and validate it.
    foundRequestStruct, apiErr := validateRequestStruct(request, endpoint);
    if (apiErr != nil) {
        return apiErr;
    }

    if (!foundRequestStruct) {
        return NewBareInternalError("-511", endpoint, "Request is not any kind of known API request.");
    }

    // Check for any special field types that we know how to populate.
    apiErr = fillRequestSpecialFields(request, endpoint);
    if (apiErr != nil) {
        return apiErr;
    }

    return nil;
}

func validateRequestStruct(request any, endpoint string) (bool, *APIError) {
    // Check all the fields (including embedded ones) for structures that we recognize as requests.
    foundRequestStruct := false;

    reflectValue := reflect.ValueOf(request).Elem();
    for i := 0; i < reflectValue.NumField(); i++ {
        fieldValue := reflectValue.Field(i);

        if (fieldValue.Type() == reflect.TypeOf((*APIRequestCourseUserContext)(nil)).Elem()) {
            // APIRequestCourseUserContext
            courseUserRequest := fieldValue.Interface().(APIRequestCourseUserContext);
            foundRequestStruct = true;

            apiErr := courseUserRequest.Validate(request, endpoint);
            if (apiErr != nil) {
                return false, apiErr;
            }

            fieldValue.Set(reflect.ValueOf(courseUserRequest));
        } else if (fieldValue.Type() == reflect.TypeOf((*APIRequestAssignmentContext)(nil)).Elem()) {
            // APIRequestAssignmentContext
            assignmentRequest := fieldValue.Interface().(APIRequestAssignmentContext);
            foundRequestStruct = true;

            apiErr := assignmentRequest.Validate(request, endpoint);
            if (apiErr != nil) {
                return false, apiErr;
            }

            fieldValue.Set(reflect.ValueOf(assignmentRequest));
        }
    }

    return foundRequestStruct, nil;
}

func fillRequestSpecialFields(request any, endpoint string) *APIError {
    reflectValue := reflect.ValueOf(request).Elem();
    structName := reflectValue.Type().Name();

    for i := 0; i < reflectValue.NumField(); i++ {
        fieldValue := reflectValue.Field(i);
        fieldType := reflectValue.Type().Field(i);

        if (fieldValue.Type() == reflect.TypeOf((*CourseUsers)(nil)).Elem()) {
            courseContextValue := reflectValue.FieldByName("APIRequestCourseUserContext");
            if (!courseContextValue.IsValid() || courseContextValue.IsZero()) {
                return NewBareInternalError("-541", endpoint, "A request with a CourseUsers field must embed APIRequestCourseUserContext").
                        Add("request", request).
                        Add("struct-name", structName).Add("field-name", fieldType.Name);
            }
            courseContext := courseContextValue.Interface().(APIRequestCourseUserContext);

            if (!fieldType.IsExported()) {
                return NewInternalError(&courseContext, "A CourseUsers field must be exported.").
                        Add("struct-name", structName).Add("field-name", fieldType.Name);
            }

            users, err := courseContext.course.GetUsers();
            if (err != nil) {
                return NewInternalError(&courseContext, "Failed to fetch embeded users.").Err(err).
                        Add("struct-name", structName).Add("field-name", fieldType.Name);
            }

            fieldValue.Set(reflect.ValueOf(users));
        }
    }

    return nil;
}

// Take a request (or any object),
// go through all the fields and look for fields typed as the encoded MinRole* fields.
// Return the maximum amongst the found roles.
// Return: (role, found role).
func getMaxRole(request any) (usr.UserRole, bool) {
    reflectValue := reflect.ValueOf(request);

    // Dereference any pointer.
    if (reflectValue.Kind() == reflect.Pointer) {
        reflectValue = reflectValue.Elem();
    }

    foundRole := false;
    role := usr.Unknown;

    for i := 0; i < reflectValue.NumField(); i++ {
        fieldValue := reflectValue.Field(i);

        if (fieldValue.Type() == reflect.TypeOf((*MinRoleOwner)(nil)).Elem()) {
            foundRole = true;
            if (role < usr.Owner) {
                role = usr.Owner;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*MinRoleAdmin)(nil)).Elem()) {
            foundRole = true;
            if (role < usr.Admin) {
                role = usr.Admin;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*MinRoleGrader)(nil)).Elem()) {
            foundRole = true;
            if (role < usr.Grader) {
                role = usr.Grader;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*MinRoleStudent)(nil)).Elem()) {
            foundRole = true;
            if (role < usr.Student) {
                role = usr.Student;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*MinRoleOther)(nil)).Elem()) {
            foundRole = true;
            if (role < usr.Other) {
                role = usr.Other;
            }
        }
    }

    return role, foundRole;
}
