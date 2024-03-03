package core

// Special fields that can be in requests.

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "reflect"

    "github.com/edulinq/autograder/config"
    "github.com/edulinq/autograder/db"
    "github.com/edulinq/autograder/model"
    "github.com/edulinq/autograder/util"
)

// The minimum user roles required encoded as a type so it can be embedded into a request struct.
type MinRoleOwner bool;
type MinRoleAdmin bool;
type MinRoleGrader bool;
type MinRoleStudent bool;
type MinRoleOther bool;

// A request having a field of this type indicates that the users for the course should be automatically fetched.
// The existence of this type in a struct also indicates that the request is at least a APIRequestCourseUserContext.
type CourseUsers map[string]*model.User;

// A request having a field of this type indicates that files from the POST request
// will be qutomatically read and written to a temp directory on disk.
type POSTFiles struct {
    TempDir string `json:"-"`
    Filenames []string `json:"-"`
}

// A request having a field of this type indicates that the request is targeting a specific user.
// This type serializes to/from a string.
// A user's email must be specified, but no error is generated if the user is not found.
// The existence of this type in a struct also indicates that the request is at least a APIRequestCourseUserContext.
type TargetUser struct {
    Found bool
    Email string
    User *model.User
}

// A request having a field of this type indicates that the request is targeting a specific user.
// This type serializes to/from a string.
// If no user is specified, then the context user is the target.
// If a user is specified, then the context user must be a grader
// (any user can acces their own resources, but higher permissions are required to access another user's resources).
// No error is generated if the user is not found.
// The existence of this type in a struct also indicates that the request is at least a APIRequestCourseUserContext.
type TargetUserSelfOrGrader struct {
    TargetUser
}

// Same as TargetUserSelfOrGrader, but for an admin context user.
type TargetUserSelfOrAdmin struct {
    TargetUser
}

// The type for a named field that must have a non-empty string value.
type NonEmptyString string;

// Check for any special request fields and validate/populate them.
func checkRequestSpecialFields(request *http.Request, apiRequest any, endpoint string) *APIError {
    reflectValue := reflect.ValueOf(apiRequest).Elem();

    for i := 0; i < reflectValue.NumField(); i++ {
        fieldValue := reflectValue.Field(i);

        if (fieldValue.Type() == reflect.TypeOf((*CourseUsers)(nil)).Elem()) {
            apiErr := checkRequestCourseUsers(endpoint, apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*TargetUser)(nil)).Elem()) {
            apiErr := checkRequestTargetUser(endpoint, apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*TargetUserSelfOrGrader)(nil)).Elem()) {
            apiErr := checkRequestTargetUserSelfOrGrader(endpoint, apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*TargetUserSelfOrAdmin)(nil)).Elem()) {
            apiErr := checkRequestTargetUserSelfOrAdmin(endpoint, apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*POSTFiles)(nil)).Elem()) {
            apiErr := checkRequestPostFiles(request, endpoint, apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        } else if (fieldValue.Type() == reflect.TypeOf((*NonEmptyString)(nil)).Elem()) {
            apiErr := checkRequestNonEmptyString(endpoint, apiRequest, i);
            if (apiErr != nil) {
                return apiErr;
            }
        }
    }

    return nil;
}

func checkRequestCourseUsers(endpoint string, apiRequest any, fieldIndex int) *APIError {
    _, users, apiErr := baseCheckRequestUsersField(endpoint, apiRequest, fieldIndex);
    if (apiErr != nil) {
        return apiErr;
    }

    reflect.ValueOf(apiRequest).Elem().Field(fieldIndex).Set(reflect.ValueOf(users));

    return nil;
}

func checkRequestTargetUser(endpoint string, apiRequest any, fieldIndex int) *APIError {
    courseContext, users, apiErr := baseCheckRequestUsersField(endpoint, apiRequest, fieldIndex);
    if (apiErr != nil) {
        return apiErr;
    }

    reflectValue := reflect.ValueOf(apiRequest).Elem();
    field := reflectValue.Field(fieldIndex).Interface().(TargetUser);

    structName := reflectValue.Type().Name();
    fieldType := reflectValue.Type().Field(fieldIndex);
    jsonName := util.JSONFieldName(fieldType);

    if (field.Email == "") {
        return NewBadRequestError("-034", &courseContext.APIRequest,
                fmt.Sprintf("Field '%s' requires a non-empty string, empty or null provided.", jsonName)).
                Add("struct-name", structName).Add("field-name", fieldType.Name).Add("json-name", jsonName);
    }

    user := users[field.Email];
    if (user == nil) {
        field.Found = false;
    } else {
        field.Found = true;
        field.User = user;
    }

    reflect.ValueOf(apiRequest).Elem().Field(fieldIndex).Set(reflect.ValueOf(field));

    return nil;
}

func checkRequestTargetUserSelfOrGrader(endpoint string, apiRequest any, fieldIndex int) *APIError {
    return checkRequestTargetUserSelfOrRole(endpoint, apiRequest, fieldIndex, model.RoleGrader);
}

func checkRequestTargetUserSelfOrAdmin(endpoint string, apiRequest any, fieldIndex int) *APIError {
    return checkRequestTargetUserSelfOrRole(endpoint, apiRequest, fieldIndex, model.RoleAdmin);
}

func checkRequestTargetUserSelfOrRole(endpoint string, apiRequest any, fieldIndex int, minRole model.UserRole) *APIError {
    courseContext, users, apiErr := baseCheckRequestUsersField(endpoint, apiRequest, fieldIndex);
    if (apiErr != nil) {
        return apiErr;
    }

    structValue := reflect.ValueOf(apiRequest).Elem().Field(fieldIndex);
    reflectField := structValue.FieldByName("TargetUser");

    field := reflectField.Interface().(TargetUser);
    if (field.Email == "") {
        field.Email = courseContext.User.Email;
    }

    // Operations not on self require higher permissions.
    if ((field.Email != courseContext.User.Email) && (courseContext.User.Role < minRole)) {
        return NewBadPermissionsError("-033", courseContext, minRole, "Non-Self Target User");
    }

    user := users[field.Email];
    if (user == nil) {
        field.Found = false;
    } else {
        field.Found = true;
        field.User = user;
    }

    reflectField.Set(reflect.ValueOf(field));

    return nil;
}

func checkRequestPostFiles(request *http.Request, endpoint string, apiRequest any, fieldIndex int) *APIError {
    reflectValue := reflect.ValueOf(apiRequest).Elem();

    structName := reflectValue.Type().Name();

    fieldValue := reflectValue.Field(fieldIndex);
    fieldType := reflectValue.Type().Field(fieldIndex);

    if (!fieldType.IsExported()) {
        return NewBareInternalError("-028", endpoint, "A POSTFiles field must be exported.").
                Add("struct-name", structName).Add("field-name", fieldType.Name);
    }

    postFiles, err := storeRequestFiles(request);

    if (err != nil) {
        switch err.(type) {
            case *fileSizeExceededError:
                return NewBareBadRequestError("-036", endpoint, err.Error()).Err(err).
                    Add("struct-name", structName).Add("field-name", fieldType.Name);
            default:
                return NewBareInternalError("-029", endpoint, "Failed to store files from POST.").Err(err).
                    Add("struct-name", structName).Add("field-name", fieldType.Name);
        }
    }

    if (postFiles == nil) {
        return NewBareBadRequestError("-030", endpoint, "Endpoint requires files to be provided in POST body, no files found.").
                Add("struct-name", structName).Add("field-name", fieldType.Name);
    }

    fieldValue.Set(reflect.ValueOf(*postFiles));

    return nil;
}

func checkRequestNonEmptyString(endpoint string, apiRequest any, fieldIndex int) *APIError {
    reflectValue := reflect.ValueOf(apiRequest).Elem();

    structName := reflectValue.Type().Name();

    fieldValue := reflectValue.Field(fieldIndex);
    fieldType := reflectValue.Type().Field(fieldIndex);
    jsonName := util.JSONFieldName(fieldType);

    value := fieldValue.Interface().(NonEmptyString);
    if (value == "") {
        return NewBareBadRequestError("-032", endpoint,
                fmt.Sprintf("Field '%s' requires a non-empty string, empty or null provided.", jsonName)).
                Add("struct-name", structName).Add("field-name", fieldType.Name).Add("json-name", jsonName);
    }

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
    inFile, fileHeader, err := request.FormFile(filename);
    if (err != nil) {
        return fmt.Errorf("Failed to access request file '%s': '%w'.", filename, err);
    }
    defer inFile.Close();

    maxFileSizeBytes := config.WEB_MAX_FILE_SIZE.Get() * 1024
    if (fileHeader.Size > int64(maxFileSizeBytes)) {
        return &fileSizeExceededError{
            Filename: filename,
            FileSize: fileHeader.Size,
            MaxFileSizeBytes: maxFileSizeBytes,
        };
    }

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

// Baseline checks for fields that require access to the course's users.
func baseCheckRequestUsersField(endpoint string, apiRequest any, fieldIndex int) (*APIRequestCourseUserContext, map[string]*model.User, *APIError) {
    reflectValue := reflect.ValueOf(apiRequest).Elem();

    fieldValue := reflectValue.Field(fieldIndex);
    fieldType := reflectValue.Type().Field(fieldIndex);

    structName := reflectValue.Type().Name();
    fieldName := fieldValue.Type().Name();

    courseContextValue := reflectValue.FieldByName("APIRequestCourseUserContext");
    if (!courseContextValue.IsValid() || courseContextValue.IsZero()) {
        return nil, nil,
            NewBareInternalError("-025", endpoint, "A request with type requiring users must embed APIRequestCourseUserContext").
                Add("request", apiRequest).
                Add("struct-name", structName).Add("field-name", fieldType.Name).Add("field-type", fieldName);
    }
    courseContext := courseContextValue.Interface().(APIRequestCourseUserContext);

    if (!fieldType.IsExported()) {
        return nil, nil,
            NewInternalError("-026", &courseContext, "Field must be exported.").
                Add("struct-name", structName).Add("field-name", fieldType.Name).Add("field-type", fieldName);
    }

    users, err := db.GetUsers(courseContext.Course);
    if (err != nil) {
        return nil, nil,
            NewInternalError("-027", &courseContext, "Failed to fetch embeded users.").Err(err).
                Add("struct-name", structName).Add("field-name", fieldType.Name).Add("field-type", fieldName);
    }

    return &courseContext, users, nil;
}

func (this *TargetUser) UnmarshalJSON(data []byte) error {
    var text string;
    err := json.Unmarshal(data, &text);
    if (err != nil) {
        return err;
    }

    if ((text == "null") || text == `""`) {
        text = "";
    }

    this.Email = text;

    return nil;
}

func (this TargetUser) MarshalJSON() ([]byte, error) {
    return json.Marshal(this.Email);
}

func (this *TargetUserSelfOrGrader) UnmarshalJSON(data []byte) error {
    return this.TargetUser.UnmarshalJSON(data);
}

func (this TargetUserSelfOrGrader) MarshalJSON() ([]byte, error) {
    return this.TargetUser.MarshalJSON();
}

func (this *TargetUserSelfOrAdmin) UnmarshalJSON(data []byte) error {
    return this.TargetUser.UnmarshalJSON(data);
}

func (this TargetUserSelfOrAdmin) MarshalJSON() ([]byte, error) {
    return this.TargetUser.MarshalJSON();
}
