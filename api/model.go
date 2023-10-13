package api

import (
    "fmt"
    "reflect"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
)

// TEST - We need validation to return objects that can be sent as responses in the case of errors.

// The minimum user roles required encoded as a type so it can be embeded into a request struct.
type MinRoleOwner bool;
type MinRoleAdmin bool;
type MinRoleGrader bool;
type MinRoleStudent bool;
type MinRoleOther bool;

// Context for a request that has a course and user (pretty much the lowest level of request).
type APIRequestCourseUserContext struct {
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

// Validate that all the fields are populated correctly and
// that they are valid in the context of this server,
// Additionally, all context fields will be populated.
// This means that this request will be authenticated here.
// The full request (object that this is embeded in) is also sent.
func (this *APIRequestCourseUserContext) Validate(request any) error {
    var err error;

    if (this.CourseID == "") {
        return fmt.Errorf("No course ID specified.");
    }

    if (this.UserEmail == "") {
        return fmt.Errorf("No user email specified.");
    }

    if (this.UserPass == "") {
        return fmt.Errorf("No user password specified.");
    }

    this.course = grader.GetCourse(this.CourseID);
    if (this.course == nil) {
        return fmt.Errorf("Could not find course '%s'.", this.CourseID);
    }

    this.user, err = AuthAPIRequest(this.course, this.UserEmail, this.UserPass);
    if (err != nil) {
        return fmt.Errorf("Failed attempt to auth user: '%w'.", err);
    }

    if (this.user == nil) {
        // TEST: Need to return unauthorized here. http.StatusUnauthorized
        return fmt.Errorf("Failed to authenticate.");
    }

    minRole, foundRole := getMaxRole(request);
    if (!foundRole) {
        return fmt.Errorf("No role found for request. All request structs require a minimum role.");
    }

    if (this.user.Role < minRole) {
        // TEST: Need to return permissions here. http.StatusForbidden
        return fmt.Errorf("Insufficient Permissions.");
    }

    return nil;
}

// See APIRequestCourseUserContext.Validate().
func (this *APIRequestAssignmentContext) Validate(request any) error {
    err := this.APIRequestCourseUserContext.Validate(request);
    if (err != nil) {
        return err;
    }

    if (this.AssignmentID == "") {
        return fmt.Errorf("No assignment ID specified.");
    }

    this.assignment = this.course.Assignments[this.AssignmentID];
    if (this.assignment == nil) {
        return fmt.Errorf("Could not find assignment '%s'.", this.AssignmentID);
    }

    return nil;
}

// Take in a pointer to an API request.
// Ensure this request has a type of known API request embedded in it and validate that embedded request.
func ValidateAPIRequest(request any) error {
    var err error;

    reflectPointer := reflect.ValueOf(request);
    if (reflectPointer.Kind() != reflect.Pointer) {
        return fmt.Errorf("ValidateAPIRequest() must be called with a pointer.");
    }

    reflectValue := reflectPointer.Elem();

    // Check all the fields (including embedded ones) for structures that we recognize as requests.
    foundRequestStruct := false;

    for i := 0; i < reflectValue.NumField(); i++ {
        fieldValue := reflectValue.Field(i);

        if (fieldValue.Type() == reflect.TypeOf((*APIRequestCourseUserContext)(nil)).Elem()) {
            // APIRequestCourseUserContext
            courseUserRequest := fieldValue.Interface().(APIRequestCourseUserContext);
            foundRequestStruct = true;

            err = courseUserRequest.Validate(request);
            if (err != nil) {
                return err;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*APIRequestAssignmentContext)(nil)).Elem()) {
            // APIRequestAssignmentContext
            assignmentRequest := fieldValue.Interface().(APIRequestAssignmentContext);
            foundRequestStruct = true;

            err = assignmentRequest.Validate(request);
            if (err != nil) {
                return err;
            }
        }
    }

    if (!foundRequestStruct) {
        return fmt.Errorf("Request is not any kind of known API request.");
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
