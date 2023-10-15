package core

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "reflect"

    "github.com/eriq-augustine/autograder/grader"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/usr"
    "github.com/eriq-augustine/autograder/util"
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

// A request having a field of this type indicates that files from the POST request
// will be qutomatically read and written to a temp directory on disk.
type POSTFiles struct {
    TempDir string `json:"-"`
    Filenames []string `json:"-"`
}

// An api request that has been reflexively verifed.
// Once validated, callers should feel safe calling reflection methods on this without extra checks.
type ValidAPIRequest any;

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
        return NewBadRequestError("-421", &this.APIRequest, "No course ID specified.");
    }

    if (this.UserEmail == "") {
        return NewBadRequestError("-422", &this.APIRequest, "No user email specified.");
    }

    if (this.UserPass == "") {
        return NewBadRequestError("-423", &this.APIRequest, "No user password specified.");
    }

    this.course = grader.GetCourse(this.CourseID);
    if (this.course == nil) {
        return NewBadRequestError("-424", &this.APIRequest, "Could not find course.").Add("course-id", this.CourseID);
    }

    this.user, apiErr = this.Auth();
    if (apiErr != nil) {
        return apiErr;
    }

    minRole, foundRole := getMaxRole(request);
    if (!foundRole) {
        return NewInternalError("-561", this, "No role found for request. All request structs require a minimum role.");
    }

    if (this.user.Role < minRole) {
        return NewBadPermissionsError("-425", this, minRole, "");
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
        return NewBadRequestError("-431", &this.APIRequest, "No assignment ID specified.");
    }

    this.assignment = this.course.Assignments[this.AssignmentID];
    if (this.assignment == nil) {
        return NewBadRequestError("-432", &this.APIRequest, "Could not find assignment.").
            Add("course-id", this.CourseID).Add("assignment-id", this.AssignmentID);
    }

    return nil;
}

// Take in a pointer to an API request.
// Ensure this request has a type of known API request embedded in it and validate that embedded request.
func ValidateAPIRequest(request *http.Request, apiRequest any, endpoint string) *APIError {
    reflectPointer := reflect.ValueOf(apiRequest);
    if (reflectPointer.Kind() != reflect.Pointer) {
        return NewBareInternalError("-511", endpoint, "ValidateAPIRequest() must be called with a pointer.");
    }

    // Ensure the request has an request type embedded, and validate it.
    foundRequestStruct, apiErr := validateRequestStruct(apiRequest, endpoint);
    if (apiErr != nil) {
        return apiErr;
    }

    if (!foundRequestStruct) {
        return NewBareInternalError("-512", endpoint, "Request is not any kind of known API request.");
    }

    // Check for any special field types that we know how to populate.
    apiErr = fillRequestSpecialFields(request, apiRequest, endpoint);
    if (apiErr != nil) {
        return apiErr;
    }

    return nil;
}

// Cleanup any resources after the response has been sent.
// This function will return an error on failure,
// but the error will generally be ignored (since this will typically be called in a defer).
// So, any error will be logged here.
func CleanupAPIrequest(apiRequest ValidAPIRequest) error {
    reflectValue := reflect.ValueOf(apiRequest).Elem();

    for i := 0; i < reflectValue.NumField(); i++ {
        fieldValue := reflectValue.Field(i);

        if (fieldValue.Type() == reflect.TypeOf((*POSTFiles)(nil)).Elem()) {
            apiErr := cleanPostFiles(apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        }
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

func fillRequestSpecialFields(request *http.Request, apiRequest any, endpoint string) *APIError {
    reflectValue := reflect.ValueOf(apiRequest).Elem();

    for i := 0; i < reflectValue.NumField(); i++ {
        fieldValue := reflectValue.Field(i);

        if (fieldValue.Type() == reflect.TypeOf((*CourseUsers)(nil)).Elem()) {
            apiErr := fillRequestCourseUsers(endpoint, apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*POSTFiles)(nil)).Elem()) {
            apiErr := fillRequestPostFiles(request, endpoint, apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        }
    }

    return nil;
}

func fillRequestCourseUsers(endpoint string, apiRequest any, fieldIndex int) *APIError {
    reflectValue := reflect.ValueOf(apiRequest).Elem();

    structName := reflectValue.Type().Name();

    fieldValue := reflectValue.Field(fieldIndex);
    fieldType := reflectValue.Type().Field(fieldIndex);

    courseContextValue := reflectValue.FieldByName("APIRequestCourseUserContext");
    if (!courseContextValue.IsValid() || courseContextValue.IsZero()) {
        return NewBareInternalError("-541", endpoint, "A request with CourseUsers must embed APIRequestCourseUserContext").
                Add("request", apiRequest).
                Add("struct-name", structName).Add("field-name", fieldType.Name);
    }
    courseContext := courseContextValue.Interface().(APIRequestCourseUserContext);

    if (!fieldType.IsExported()) {
        return NewInternalError("-542", &courseContext, "A CourseUsers field must be exported.").
                Add("struct-name", structName).Add("field-name", fieldType.Name);
    }

    users, err := courseContext.course.GetUsers();
    if (err != nil) {
        return NewInternalError("-543", &courseContext, "Failed to fetch embeded users.").Err(err).
                Add("struct-name", structName).Add("field-name", fieldType.Name);
    }

    fieldValue.Set(reflect.ValueOf(users));

    return nil;
}

func fillRequestPostFiles(request *http.Request, endpoint string, apiRequest any, fieldIndex int) *APIError {
    reflectValue := reflect.ValueOf(apiRequest).Elem();

    structName := reflectValue.Type().Name();

    fieldValue := reflectValue.Field(fieldIndex);
    fieldType := reflectValue.Type().Field(fieldIndex);

    if (!fieldType.IsExported()) {
        return NewBareInternalError("-551", endpoint, "A POSTFiles field must be exported.").
                Add("struct-name", structName).Add("field-name", fieldType.Name);
    }

    postFiles, err := storeRequestFiles(request);

    if (err != nil) {
        return NewBareInternalError("-552", endpoint, "Failed to store files from POST.").Err(err).
                Add("struct-name", structName).Add("field-name", fieldType.Name);
    }

    if (postFiles == nil) {
        return NewBareBadRequestError("-411", endpoint, "Endpoint requires files to be provided in POST body, no files found.").
                Add("struct-name", structName).Add("field-name", fieldType.Name);
    }

    fieldValue.Set(reflect.ValueOf(*postFiles));

    return nil;
}

func cleanPostFiles(apiRequest ValidAPIRequest, fieldIndex int) *APIError {
    reflectValue := reflect.ValueOf(apiRequest).Elem();
    fieldValue := reflectValue.Field(fieldIndex);
    postFiles := fieldValue.Interface().(POSTFiles);
    os.RemoveAll(postFiles.TempDir);

    return nil;
}

func storeRequestFiles(request *http.Request) (*POSTFiles, error) {
    if (request.MultipartForm == nil) {
        return nil, nil;
    }

    if (len(request.MultipartForm.File) == 0) {
        return nil, nil;
    }

    tempDir, err := util.MkDirTemp("api-request-files-");
    if (err != nil) {
        return nil, fmt.Errorf("Failed to create temp api files directory: '%w'.", err);
    }

    filenames := make([]string, 0, len(request.MultipartForm.File));

    // Use an inner function to help control the removal of the temp dir on error.
    innerFunc := func() error {
        for filename, _ := range request.MultipartForm.File {
            filenames = append(filenames, filename);

            err = storeRequestFile(request, tempDir, filename);
            if (err != nil) {
                return err;
            }
        }

        return nil;
    }

    err = innerFunc();
    if (err != nil) {
        os.RemoveAll(tempDir);
        return nil, err;
    }

    postFiles := POSTFiles{
        TempDir: tempDir,
        Filenames: filenames,
    };

    return &postFiles, nil;
}

func storeRequestFile(request *http.Request, outDir string, filename string) error {
    inFile, _, err := request.FormFile(filename);
    if (err != nil) {
        return fmt.Errorf("Failed to access request file '%s': '%w'.", filename, err);
    }
    defer inFile.Close();

    outPath := filepath.Join(outDir, filename);

    outFile, err := os.Create(outPath);
    if (err != nil) {
        return fmt.Errorf("Failed to create output file '%s': '%w'.", outPath, err);
    }
    defer outFile.Close();

    _, err = io.Copy(outFile, inFile);
    if (err != nil) {
        return fmt.Errorf("Failed to copy contents of request file '%s': '%w'.", filename, err);
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
