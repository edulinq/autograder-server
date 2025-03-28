package core

// Special fields that can be in requests.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// A request having a field of this type indicates that the users for the course should be automatically fetched.
// The existence of this type in a struct also indicates that the request is at least a APIRequestCourseUserContext.
type CourseUsers map[string]*model.CourseUser

// A request having a field of this type indicates that files from the POST request
// will be qutomatically read and written to a temp directory on disk.
type POSTFiles struct {
	TempDir   string   `json:"-"`
	Filenames []string `json:"-"`
}

// A request having a field of this type indicates that the request is targeting a specific server user.
// This type serializes to/from a string.
// A user's email must be specified, but no error is generated if the user is not found.
// The existence of this type in a struct also indicates that the request is at least a APIRequestUserContext.
type TargetServerUser struct {
	Found bool
	Email string
	User  *model.ServerUser
}

// A request having a field of this type indicates that the request is targeting a specific course user.
// This type serializes to/from a string.
// A user's email must be specified, but no error is generated if the user is not found.
// The existence of this type in a struct also indicates that the request is at least a APIRequestCourseUserContext.
type TargetCourseUser struct {
	Found bool
	Email string
	User  *model.CourseUser
}

// A request having a field of this type indicates that the request is targeting a specific user.
// This type serializes to/from a string.
// If no user is specified, then the context user is the target.
// If a user is specified, then the context user must be a server admin
// (any user can access their own resources, but higher permissions are required to access another user's resources).
// No error is generated if the user is not found.
// The existence of this type in a struct also indicates that the request is at least a APIRequestUserContext.
type TargetServerUserSelfOrAdmin struct {
	TargetServerUser
}

// Same as TargetServerUserSelfOrAdmin, but in the context of a course user and a grader context user.
// Therefore, the context user only has to be a grader in the context course (or the target user themself).
// When targeting yourself, the user can be a server admin (and will be escalated to course owner for the request).
// The existence of this type in a struct also indicates that the request is at least a APIRequestCourseUserContext.
type TargetCourseUserSelfOrGrader struct {
	TargetCourseUser
}

// Same as TargetCourseUserSelfOrGrader, but for a course admin context user.
type TargetCourseUserSelfOrAdmin struct {
	TargetCourseUser
}

// The type for a named field that must have a non-empty string value.
type NonEmptyString string

// Check for any special request fields and validate/populate them.
func checkRequestSpecialFields(request *http.Request, apiRequest any, endpoint string) *APIError {
	reflectValue := reflect.ValueOf(apiRequest).Elem()

	var apiErr *APIError = nil
	var postFiles *POSTFiles = nil

	// Use a func to better handle cleanup on error.
	innerFunc := func() *APIError {
		for i := 0; i < reflectValue.NumField(); i++ {
			fieldValue := reflectValue.Field(i)

			if fieldValue.Type() == reflect.TypeOf((*CourseUsers)(nil)).Elem() {
				apiErr = checkRequestCourseUsers(endpoint, apiRequest, i)
				if apiErr != nil {
					return apiErr
				}
			} else if fieldValue.Type() == reflect.TypeOf((*TargetServerUser)(nil)).Elem() {
				apiErr = checkRequestTargetServerUser(endpoint, apiRequest, i)
				if apiErr != nil {
					return apiErr
				}
			} else if fieldValue.Type() == reflect.TypeOf((*TargetCourseUser)(nil)).Elem() {
				apiErr = checkRequestTargetCourseUser(endpoint, apiRequest, i)
				if apiErr != nil {
					return apiErr
				}
			} else if fieldValue.Type() == reflect.TypeOf((*TargetServerUserSelfOrAdmin)(nil)).Elem() {
				apiErr = checkRequestTargetServerUserSelfOrAdmin(endpoint, apiRequest, i)
				if apiErr != nil {
					return apiErr
				}
			} else if fieldValue.Type() == reflect.TypeOf((*TargetCourseUserSelfOrGrader)(nil)).Elem() {
				apiErr = checkRequestTargetCourseUserSelfOrGrader(endpoint, apiRequest, i)
				if apiErr != nil {
					return apiErr
				}
			} else if fieldValue.Type() == reflect.TypeOf((*TargetCourseUserSelfOrAdmin)(nil)).Elem() {
				apiErr = checkRequestTargetCourseUserSelfOrAdmin(endpoint, apiRequest, i)
				if apiErr != nil {
					return apiErr
				}
			} else if fieldValue.Type() == reflect.TypeOf((*POSTFiles)(nil)).Elem() {
				postFiles, apiErr = checkRequestPostFiles(request, endpoint, apiRequest, i)
				if apiErr != nil {
					return apiErr
				}
			} else if fieldValue.Type() == reflect.TypeOf((*NonEmptyString)(nil)).Elem() {
				apiErr = checkRequestNonEmptyString(endpoint, apiRequest, i)
				if apiErr != nil {
					return apiErr
				}
			}
		}

		return nil
	}

	err := innerFunc()
	if err != nil {
		if postFiles != nil {
			util.RemoveDirent(postFiles.TempDir)
		}

		return err
	}

	return nil
}

func checkRequestCourseUsers(endpoint string, apiRequest any, fieldIndex int) *APIError {
	_, users, apiErr := baseCheckRequestUsersField(endpoint, apiRequest, fieldIndex)
	if apiErr != nil {
		return apiErr
	}

	reflect.ValueOf(apiRequest).Elem().Field(fieldIndex).Set(reflect.ValueOf(users))

	return nil
}

// The base checks for any TargetServerUser* field.
func baseCheckRequestTargetServerUser(endpoint string, apiRequest any, fieldIndex int) (*APIRequestUserContext, *APIError) {
	reflectValue := reflect.ValueOf(apiRequest).Elem()

	fieldValue := reflectValue.Field(fieldIndex)
	fieldType := reflectValue.Type().Field(fieldIndex)

	structName := reflectValue.Type().Name()
	fieldName := fieldValue.Type().Name()

	// Check the request.

	userContextValue := reflectValue.FieldByName("APIRequestUserContext")
	if !userContextValue.IsValid() || userContextValue.IsZero() {
		return nil, NewBareInternalError("-042", endpoint, "A request with type requiring a server user must embed APIRequestUserContext").
			Add("request", apiRequest).
			Add("struct-name", structName).Add("field-name", fieldType.Name).Add("field-type", fieldName)
	}
	userContext := userContextValue.Interface().(APIRequestUserContext)

	if !fieldType.IsExported() {
		return nil, NewUserContextInternalError("-043", &userContext, "Field must be exported.").
			Add("struct-name", structName).Add("field-name", fieldType.Name).Add("field-type", fieldName)
	}

	return &userContext, nil
}

// Check a TargetServerUser field.
func checkRequestTargetServerUser(endpoint string, apiRequest any, fieldIndex int) *APIError {
	userContext, apiErr := baseCheckRequestTargetServerUser(endpoint, apiRequest, fieldIndex)
	if apiErr != nil {
		return apiErr
	}

	reflectValue := reflect.ValueOf(apiRequest).Elem()
	field := reflectValue.Field(fieldIndex).Interface().(TargetServerUser)

	structName := reflectValue.Type().Name()
	fieldType := reflectValue.Type().Field(fieldIndex)
	jsonName := util.JSONFieldName(fieldType)

	if field.Email == "" {
		return NewBadRequestError("-044", &userContext.APIRequest,
			fmt.Sprintf("Field '%s' requires a non-empty string, empty or null provided.", jsonName)).
			Add("struct-name", structName).Add("field-name", fieldType.Name).Add("json-name", jsonName)
	}

	user, err := db.GetServerUser(field.Email)
	if err != nil {
		return NewUserContextInternalError("-045", userContext, "Failed to fetch user from DB.").
			Add("email", field.Email).Err(err)
	}

	if user == nil {
		field.Found = false
	} else {
		field.Found = true
		field.User = user
	}

	reflect.ValueOf(apiRequest).Elem().Field(fieldIndex).Set(reflect.ValueOf(field))

	return nil
}

func checkRequestTargetServerUserSelfOrAdmin(endpoint string, apiRequest any, fieldIndex int) *APIError {
	return checkRequestTargetServerUserSelfOrRole(endpoint, apiRequest, fieldIndex, model.ServerRoleAdmin)
}

func checkRequestTargetServerUserSelfOrRole(endpoint string, apiRequest any, fieldIndex int, minRole model.ServerUserRole) *APIError {
	userContext, apiErr := baseCheckRequestTargetServerUser(endpoint, apiRequest, fieldIndex)
	if apiErr != nil {
		return apiErr
	}

	structValue := reflect.ValueOf(apiRequest).Elem().Field(fieldIndex)
	reflectField := structValue.FieldByName("TargetServerUser")

	field := reflectField.Interface().(TargetServerUser)
	if field.Email == "" {
		field.Email = userContext.ServerUser.Email
	}

	// Operations not on self require higher permissions.
	if (field.Email != userContext.ServerUser.Email) && (userContext.ServerUser.Role < minRole) {
		return NewBadServerPermissionsError("-046", userContext, minRole, "Non-Self Target User")
	}

	user, err := db.GetServerUser(field.Email)
	if err != nil {
		return NewUserContextInternalError("-047", userContext, "Failed to fetch user from DB.").
			Add("email", field.Email).Err(err)
	}

	if user == nil {
		field.Found = false
	} else {
		field.Found = true
		field.User = user
	}

	reflectField.Set(reflect.ValueOf(field))

	return nil
}

func checkRequestTargetCourseUser(endpoint string, apiRequest any, fieldIndex int) *APIError {
	courseContext, users, apiErr := baseCheckRequestUsersField(endpoint, apiRequest, fieldIndex)
	if apiErr != nil {
		return apiErr
	}

	reflectValue := reflect.ValueOf(apiRequest).Elem()
	field := reflectValue.Field(fieldIndex).Interface().(TargetCourseUser)

	structName := reflectValue.Type().Name()
	fieldType := reflectValue.Type().Field(fieldIndex)
	jsonName := util.JSONFieldName(fieldType)

	if field.Email == "" {
		return NewBadRequestError("-034", &courseContext.APIRequest,
			fmt.Sprintf("Field '%s' requires a non-empty string, empty or null provided.", jsonName)).
			Add("struct-name", structName).Add("field-name", fieldType.Name).Add("json-name", jsonName)
	}

	user := users[field.Email]
	if user == nil {
		field.Found = false
	} else {
		field.Found = true
		field.User = user
	}

	reflect.ValueOf(apiRequest).Elem().Field(fieldIndex).Set(reflect.ValueOf(field))

	return nil
}

func checkRequestTargetCourseUserSelfOrGrader(endpoint string, apiRequest any, fieldIndex int) *APIError {
	return checkRequestTargetCourseUserSelfOrRole(endpoint, apiRequest, fieldIndex, model.CourseRoleGrader)
}

func checkRequestTargetCourseUserSelfOrAdmin(endpoint string, apiRequest any, fieldIndex int) *APIError {
	return checkRequestTargetCourseUserSelfOrRole(endpoint, apiRequest, fieldIndex, model.CourseRoleAdmin)
}

func checkRequestTargetCourseUserSelfOrRole(endpoint string, apiRequest any, fieldIndex int, minRole model.CourseUserRole) *APIError {
	courseContext, users, apiErr := baseCheckRequestUsersField(endpoint, apiRequest, fieldIndex)
	if apiErr != nil {
		return apiErr
	}

	structValue := reflect.ValueOf(apiRequest).Elem().Field(fieldIndex)
	reflectField := structValue.FieldByName("TargetCourseUser")

	field := reflectField.Interface().(TargetCourseUser)
	if field.Email == "" {
		field.Email = courseContext.User.Email

		// If the user (which is the target of this request) is a server admin,
		// insert them into the course users (they have already been escalated).
		users[courseContext.User.Email] = courseContext.User
	}

	// Operations not on self require higher permissions.
	if (field.Email != courseContext.User.Email) && (courseContext.User.Role < minRole) {
		return NewBadCoursePermissionsError("-033", courseContext, minRole, "Non-Self Target User")
	}

	user := users[field.Email]
	if user == nil {
		field.Found = false
	} else {
		field.Found = true
		field.User = user
	}

	reflectField.Set(reflect.ValueOf(field))

	return nil
}

// Get files from the POST request.
// If successful, a pointer to the post files will be returned AND embedded in the request.
// The returned reference is for cleanup if there is an error further up the stack,
// if there is no error it does not need to be explicitly cleaned up (as that will happen after the request handler is done).
func checkRequestPostFiles(request *http.Request, endpoint string, apiRequest any, fieldIndex int) (*POSTFiles, *APIError) {
	reflectValue := reflect.ValueOf(apiRequest).Elem()

	structName := reflectValue.Type().Name()

	fieldValue := reflectValue.Field(fieldIndex)
	fieldType := reflectValue.Type().Field(fieldIndex)

	if !fieldType.IsExported() {
		return nil, NewBareInternalError("-028", endpoint, "A POSTFiles field must be exported.").
			Add("struct-name", structName).Add("field-name", fieldType.Name)
	}

	postFiles, err := storeRequestFiles(request)

	if err != nil {
		var fileSizeExceededError *fileSizeExceededError
		if errors.As(err, &fileSizeExceededError) {
			return nil, NewBareBadRequestError("-036", endpoint, err.Error()).Err(err).
				Add("struct-name", structName).Add("field-name", fieldType.Name).
				Add("filename", fileSizeExceededError.Filename).Add("file-size", fileSizeExceededError.FileSizeKB).
				Add("max-file-size-kb", fileSizeExceededError.MaxFileSizeKB)
		} else {
			return nil, NewBareInternalError("-029", endpoint, "Failed to store files from POST.").Err(err).
				Add("struct-name", structName).Add("field-name", fieldType.Name)
		}
	}

	if postFiles == nil {
		return nil, NewBareBadRequestError("-030", endpoint, "Endpoint requires files to be provided in POST body, no files found.").
			Add("struct-name", structName).Add("field-name", fieldType.Name)
	}

	fieldValue.Set(reflect.ValueOf(*postFiles))

	return postFiles, nil
}

func checkRequestNonEmptyString(endpoint string, apiRequest any, fieldIndex int) *APIError {
	reflectValue := reflect.ValueOf(apiRequest).Elem()

	structName := reflectValue.Type().Name()

	fieldValue := reflectValue.Field(fieldIndex)
	fieldType := reflectValue.Type().Field(fieldIndex)
	jsonName := util.JSONFieldName(fieldType)

	value := fieldValue.Interface().(NonEmptyString)
	if value == "" {
		return NewBareBadRequestError("-038", endpoint,
			fmt.Sprintf("Field '%s' requires a non-empty string, empty or null provided.", jsonName)).
			Add("struct-name", structName).Add("field-name", fieldType.Name).Add("json-name", jsonName)
	}

	return nil
}

func cleanPostFiles(apiRequest ValidAPIRequest, fieldIndex int) *APIError {
	reflectValue := reflect.ValueOf(apiRequest).Elem()
	fieldValue := reflectValue.Field(fieldIndex)
	postFiles := fieldValue.Interface().(POSTFiles)
	util.RemoveDirent(postFiles.TempDir)

	return nil
}

func storeRequestFiles(request *http.Request) (*POSTFiles, error) {
	if request.MultipartForm == nil {
		return nil, nil
	}

	if len(request.MultipartForm.File) == 0 {
		return nil, nil
	}

	tempDir, err := util.MkDirTemp("api-request-files-")
	if err != nil {
		return nil, fmt.Errorf("Failed to create temp api files directory: '%w'.", err)
	}

	filenames := make([]string, 0, len(request.MultipartForm.File))

	// Use an inner function to help control the removal of the temp dir on error.
	innerFunc := func() error {
		for filename, _ := range request.MultipartForm.File {
			filenames = append(filenames, filename)

			err = storeRequestFile(request, tempDir, filename)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = innerFunc()
	if err != nil {
		util.RemoveDirent(tempDir)
		return nil, err
	}

	postFiles := POSTFiles{
		TempDir:   tempDir,
		Filenames: filenames,
	}

	return &postFiles, nil
}

func storeRequestFile(request *http.Request, outDir string, filename string) error {
	inFile, fileHeader, err := request.FormFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to access request file '%s': '%w'.", filename, err)
	}
	defer inFile.Close()

	maxFileSizeKB := int64(config.WEB_MAX_FILE_SIZE_KB.Get())
	if fileHeader.Size > maxFileSizeKB*1024 {
		return &fileSizeExceededError{
			Filename:      filename,
			FileSizeKB:    fileHeader.Size / 1024,
			MaxFileSizeKB: maxFileSizeKB,
		}
	}

	outPath := filepath.Join(outDir, filename)

	err = util.WriteFileFromReader(outPath, inFile)
	if err != nil {
		return fmt.Errorf("Failed to store copy of request file '%s': '%w'.", outPath, err)
	}

	return nil
}

// Baseline checks for fields that require access to the course's users.
func baseCheckRequestUsersField(endpoint string, apiRequest any, fieldIndex int) (*APIRequestCourseUserContext, map[string]*model.CourseUser, *APIError) {
	reflectValue := reflect.ValueOf(apiRequest).Elem()

	fieldValue := reflectValue.Field(fieldIndex)
	fieldType := reflectValue.Type().Field(fieldIndex)

	structName := reflectValue.Type().Name()
	fieldName := fieldValue.Type().Name()

	courseContextValue := reflectValue.FieldByName("APIRequestCourseUserContext")
	if !courseContextValue.IsValid() || courseContextValue.IsZero() {
		return nil, nil,
			NewBareInternalError("-025", endpoint, "A request with type requiring course users must embed APIRequestCourseUserContext").
				Add("request", apiRequest).
				Add("struct-name", structName).Add("field-name", fieldType.Name).Add("field-type", fieldName)
	}
	courseContext := courseContextValue.Interface().(APIRequestCourseUserContext)

	if !fieldType.IsExported() {
		return nil, nil,
			NewInternalError("-026", &courseContext, "Field must be exported.").
				Add("struct-name", structName).Add("field-name", fieldType.Name).Add("field-type", fieldName)
	}

	users, err := db.GetCourseUsers(courseContext.Course)
	if err != nil {
		return nil, nil,
			NewInternalError("-027", &courseContext, "Failed to fetch embeded users.").Err(err).
				Add("struct-name", structName).Add("field-name", fieldType.Name).Add("field-type", fieldName)
	}

	return &courseContext, users, nil
}

func (this *TargetServerUser) UnmarshalJSON(data []byte) error {
	var text string
	err := json.Unmarshal(data, &text)
	if err != nil {
		return err
	}

	if (text == "null") || text == `""` {
		text = ""
	}

	this.Email = text

	return nil
}

func (this TargetServerUser) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.Email)
}

func (this *TargetServerUserSelfOrAdmin) UnmarshalJSON(data []byte) error {
	return this.TargetServerUser.UnmarshalJSON(data)
}

func (this TargetServerUserSelfOrAdmin) MarshalJSON() ([]byte, error) {
	return this.TargetServerUser.MarshalJSON()
}

func (this *TargetCourseUser) UnmarshalJSON(data []byte) error {
	var text string
	err := json.Unmarshal(data, &text)
	if err != nil {
		return err
	}

	if (text == "null") || text == `""` {
		text = ""
	}

	this.Email = text

	return nil
}

func (this TargetCourseUser) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.Email)
}

func (this *TargetCourseUserSelfOrGrader) UnmarshalJSON(data []byte) error {
	return this.TargetCourseUser.UnmarshalJSON(data)
}

func (this TargetCourseUserSelfOrGrader) MarshalJSON() ([]byte, error) {
	return this.TargetCourseUser.MarshalJSON()
}

func (this *TargetCourseUserSelfOrAdmin) UnmarshalJSON(data []byte) error {
	return this.TargetCourseUser.UnmarshalJSON(data)
}

func (this TargetCourseUserSelfOrAdmin) MarshalJSON() ([]byte, error) {
	return this.TargetCourseUser.MarshalJSON()
}

// A special error for when a submitted file exceeds the defined maximum allowable size.
type fileSizeExceededError struct {
	Filename      string
	FileSizeKB    int64
	MaxFileSizeKB int64
}

func (this *fileSizeExceededError) Error() string {
	return fmt.Sprintf("File '%s' is %d KB. The maximum allowable size is %d KB.", this.Filename, this.FileSizeKB, this.MaxFileSizeKB)
}
