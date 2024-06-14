package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

// Test CourseUsers, TargetUserSelfOrGrader, and TargetUserSelfOrAdmin.
// No embeded course context.
func TestBadUsersFieldNoContext(test *testing.T) {
	testCases := []struct{ request any }{
		{&struct{ Users CourseUsers }{}},
		{&struct{ User TargetUserSelfOrGrader }{}},
		{&struct{ User TargetUserSelfOrAdmin }{}},
	}

	for i, testCase := range testCases {
		apiErr := checkRequestSpecialFields(nil, testCase.request, "")
		if apiErr == nil {
			test.Fatalf("Case %d: Struct with no course context does not return an error: '%+v'.",
				i, testCase.request)
		}

		if apiErr.Locator != "-025" {
			test.Fatalf("Case %d: Struct with no course context does not return an error with locator '-025', found '%s': '%+v.",
				i, apiErr.Locator, testCase.request)
		}
	}
}

// Test CourseUsers, TargetUserSelfOrGrader, and TargetUserSelfOrAdmin.
// Users are not exported.
func TestBadUsersFieldNotExported(test *testing.T) {
	testCases := []struct{ request any }{
		{
			&struct {
				APIRequestCourseUserContext
				MinCourseRoleStudent
				users CourseUsers
			}{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "student@test.com",
						UserPass:  studentPass,
					},
					CourseID: "course101",
				},
			},
		},
		{
			&struct {
				APIRequestCourseUserContext
				MinCourseRoleStudent
				targetUser TargetUserSelfOrGrader
			}{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "student@test.com",
						UserPass:  studentPass,
					},
					CourseID: "course101",
				},
			},
		},
		{
			&struct {
				APIRequestCourseUserContext
				MinCourseRoleStudent
				targetUser TargetUserSelfOrAdmin
			}{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "student@test.com",
						UserPass:  studentPass,
					},
					CourseID: "course101",
				},
			},
		},
	}

	for i, testCase := range testCases {
		apiErr := ValidateAPIRequest(nil, testCase.request, "")
		if apiErr == nil {
			test.Fatalf("Case %d: Struct with non-exported field does not return an error: '%+v'.",
				i, testCase.request)
		}

		if apiErr.Locator != "-026" {
			test.Fatalf("Case %d: Struct with non-exported field does not return an error with locator '-026', found '%s': '%v.",
				i, apiErr.Locator, apiErr)
		}
	}
}

func TestNonEmptyStringField(test *testing.T) {
	testCases := []struct {
		request  any
		errLoc   string
		jsonName string
	}{
		{&struct {
			APIRequest
			Text string
		}{}, "", ""},

		{&struct {
			APIRequest
			Text NonEmptyString
		}{Text: "ZZZ"}, "", "Text"},

		{&struct {
			APIRequest
			Text NonEmptyString
		}{}, "-038", "Text"},
		{&struct {
			APIRequest
			Text NonEmptyString
		}{Text: ""}, "-038", "Text"},

		{&struct {
			APIRequest
			Text NonEmptyString `json:"text"`
		}{}, "-038", "text"},
		{&struct {
			APIRequest
			Text NonEmptyString `json:"text,omitempty"`
		}{}, "-038", "text"},
		{&struct {
			APIRequest
			Text NonEmptyString `json:"foo-bar"`
		}{}, "-038", "foo-bar"},
		{&struct {
			APIRequest
			Text NonEmptyString `json:"foo-bar,omitempty"`
		}{}, "-038", "foo-bar"},
	}

	for i, testCase := range testCases {
		apiErr := ValidateAPIRequest(nil, testCase.request, "")
		if apiErr != nil {
			if testCase.errLoc != "" {
				if testCase.errLoc != apiErr.Locator {
					test.Errorf("Case %d: Incorrect error returned on empty string. Expcted '%s', found '%s'.",
						i, testCase.errLoc, apiErr.Locator)
				} else {
					if testCase.jsonName != apiErr.AdditionalDetails["json-name"] {
						test.Errorf("Case %d: Incorrect JSON name returned. Expcted '%s', found '%s'.",
							i, testCase.jsonName, apiErr.AdditionalDetails["json-name"])
					}
				}
			} else {
				test.Errorf("Case %d: Error retutned when it should not be: '%v'.", i, apiErr)
			}
		} else {
			if testCase.errLoc != "" {
				test.Errorf("Case %d: Error not retutned when it should be.", i)
			}
		}
	}
}

func TestGoodPostFiles(test *testing.T) {
	endpoint := `/test/api/post-files/good`

	type requestType struct {
		APIRequestCourseUserContext
		MinCourseRoleStudent

		Files POSTFiles
	}

	var tempDir string

	handler := func(request *requestType) (*string, *APIError) {
		if len(request.Files.Filenames) != 1 {
			response := fmt.Sprintf("Incorrect number of files. Expected 1, got '%d'.", len(request.Files.Filenames))
			return &response, nil
		}

		path := filepath.Join(request.Files.TempDir, request.Files.Filenames[0])
		text, err := util.ReadFile(path)
		if err != nil {
			response := fmt.Sprintf("Unable to get files contents from '%s': '%v'.", path, err)
			return &response, nil
		}

		text = strings.TrimSpace(text)

		expectedText := "a"
		if text != expectedText {
			response := fmt.Sprintf("File text not as expected. Expected: '%s', actual: '%s'.", expectedText, text)
			return &response, nil
		}

		tempDir = request.Files.TempDir

		return nil, nil
	}

	routes = append(routes, NewAPIRoute(endpoint, handler))

	paths := []string{
		filepath.Join(util.RootDirForTesting(), "testdata", "files", "a.txt"),
	}

	response := SendTestAPIRequestFull(test, endpoint, nil, paths, model.CourseRoleAdmin)
	if response.Content != nil {
		test.Fatalf("Handler gave an error: '%s'.", response.Content)
	}

	// Check that the temp dir was cleaned up.
	if util.PathExists(tempDir) {
		test.Fatalf("Temp dir was not cleaned up: '%s'.", tempDir)
	}
}

func TestBadPostFilesFieldNotExported(test *testing.T) {
	// Files are not exported.
	type badRequestType struct {
		APIRequestCourseUserContext
		MinCourseRoleStudent

		files POSTFiles
	}

	request := badRequestType{
		APIRequestCourseUserContext: standardCourseContext,
	}

	apiErr := ValidateAPIRequest(nil, &request, "")
	if apiErr == nil {
		test.Fatalf("Struct with non-exported files does not return an error,")
	}

	expectedLocator := "-028"
	if apiErr.Locator != expectedLocator {
		test.Fatalf("Struct with non-exported files does not return an error with the correct locator. Expcted '%s', found '%s'.",
			expectedLocator, apiErr.Locator)
	}
}

func TestBadPostFilesNoFiles(test *testing.T) {
	endpoint := `/test/api/post-files/bad/no-files`

	type requestType struct {
		APIRequestCourseUserContext
		MinCourseRoleStudent

		Files POSTFiles
	}

	handler := func(request *requestType) (*any, *APIError) {
		return nil, nil
	}

	routes = append(routes, NewAPIRoute(endpoint, handler))

	paths := []string{}

	response := SendTestAPIRequestFull(test, endpoint, nil, paths, model.CourseRoleAdmin)
	if response.Success {
		test.Fatalf("Request did not generate an error: '%v'.", response)
	}

	expectedLocator := "-030"
	if response.Locator != expectedLocator {
		test.Fatalf("Error does not have the correct locator. Expcted '%s', found '%s'.",
			expectedLocator, response.Locator)
	}
}

func TestBadPostFilesStoreFail(test *testing.T) {
	endpoint := `/test/api/post-files/bad/store-fail`

	type requestType struct {
		APIRequestCourseUserContext
		MinCourseRoleStudent

		Files POSTFiles
	}

	handler := func(request *requestType) (*any, *APIError) {
		return nil, nil
	}

	routes = append(routes, NewAPIRoute(endpoint, handler))

	paths := []string{
		filepath.Join(util.RootDirForTesting(), "testdata", "files", "a.txt"),
	}

	// Ensure that storing the files will fail.
	util.SetTempDirForTesting(os.DevNull)
	defer util.SetTempDirForTesting("")

	response := SendTestAPIRequestFull(test, endpoint, nil, paths, model.CourseRoleAdmin)
	if response.Success {
		test.Fatalf("Request did not generate an error: '%v'.", response)
	}

	expectedLocator := "-029"
	if response.Locator != expectedLocator {
		test.Fatalf("Error does not have the correct locator. Expcted '%s', found '%s'.",
			expectedLocator, response.Locator)
	}
}

func TestBadPostFilesFileSizeExceeded(test *testing.T) {
	resetMaxFileSize := config.WEB_MAX_FILE_SIZE_KB.Get()

	// Set size to 1 KB for testing, then reset when done testing.
	config.WEB_MAX_FILE_SIZE_KB.Set(1)
	defer config.WEB_MAX_FILE_SIZE_KB.Set(resetMaxFileSize)

	endpoint := `/test/api/post-files/bad/size-exceeded`

	type requestType struct {
		APIRequestCourseUserContext
		MinCourseRoleStudent

		Files POSTFiles
	}

	handler := func(request *requestType) (*any, *APIError) {
		return nil, nil
	}

	routes = append(routes, NewAPIRoute(endpoint, handler))

	// Two paths provided: a.txt is under the size limit, 1092bytes.txt is over the size limit.
	paths := []string{
		filepath.Join(util.RootDirForTesting(), "testdata", "files", "1092bytes.txt"),
	}

	response := SendTestAPIRequestFull(test, endpoint, nil, paths, model.CourseRoleAdmin)
	if response.Success {
		test.Fatalf("Request did not generate an error: '%v'.", response)
	}

	expectedLocator := "-036"
	if response.Locator != expectedLocator {
		test.Fatalf("Error does not have the correct locator. Expected '%s', found '%s'.",
			expectedLocator, response.Locator)
	}
}

func TestTargetUserJSON(test *testing.T) {
	createTargetType := func(targetUser TargetUser) TargetUser {
		return targetUser
	}

	testTargetUserJSON(test, createTargetType)
}

func TestTargetUserSelfOrGraderJSON(test *testing.T) {
	createTargetType := func(targetUser TargetUser) TargetUserSelfOrGrader {
		return TargetUserSelfOrGrader{targetUser}
	}

	testTargetUserJSON(test, createTargetType)
}

func TestTargetUserSelfOrAdminJSON(test *testing.T) {
	createTargetType := func(targetUser TargetUser) TargetUserSelfOrAdmin {
		return TargetUserSelfOrAdmin{targetUser}
	}

	testTargetUserJSON(test, createTargetType)
}

func testTargetUserJSON[T comparable](test *testing.T, createTargetType func(TargetUser) T) {
	testCases := []struct {
		in       string
		expected T
	}{
		{`""`, createTargetType(TargetUser{false, "", nil})},
		{`"a"`, createTargetType(TargetUser{false, "a", nil})},
		{`"student@test.com"`, createTargetType(TargetUser{false, "student@test.com", nil})},
		{`"a\"b\"c"`, createTargetType(TargetUser{false, `a"b"c`, nil})},
	}

	for i, testCase := range testCases {
		var target T
		err := json.Unmarshal([]byte(testCase.in), &target)
		if err != nil {
			test.Errorf("Case %d: Failed to unmarshal: '%v'.", i, err)
			continue
		}

		if testCase.expected != target {
			test.Errorf("Case %d: Result not as expected. Expected: '%+v', Actual: '%+v'.", i, testCase.expected, target)
			continue
		}

		out, err := util.ToJSON(target)
		if err != nil {
			test.Errorf("Case %d: Failed to marshal: '%v'.", i, err)
			continue
		}

		if testCase.in != out {
			test.Errorf("Case %d: Remarshal does not produce the same as input. Expected: '%+v', Actual: '%+v'.", i, testCase.in, out)
			continue
		}
	}
}

func TestTargetUserSelfOrGrader(test *testing.T) {
	createTargetType := func(targetUser TargetUser) TargetUserSelfOrGrader {
		return TargetUserSelfOrGrader{targetUser}
	}

	createRequest := func(role model.CourseUserRole, target string) *testTargetUserSelfOrGraderRequestType {
		return &testTargetUserSelfOrGraderRequestType{
			APIRequestCourseUserContext: APIRequestCourseUserContext{
				APIRequestUserContext: APIRequestUserContext{
					UserEmail: role.String() + "@test.com",
					UserPass:  util.Sha256HexFromString(role.String()),
				},
				CourseID: "course101",
			},
			User: TargetUserSelfOrGrader{
				TargetUser{
					Email: target,
				},
			},
		}
	}

	isNonSelfPermError := func(role model.CourseUserRole) bool {
		return role < model.CourseRoleGrader
	}

	testTargetUser(test, createTargetType, createRequest, isNonSelfPermError)
}

type testTargetUserSelfOrGraderRequestType struct {
	APIRequestCourseUserContext
	MinCourseRoleOther

	User TargetUserSelfOrGrader
}

func (this *testTargetUserSelfOrGraderRequestType) GetUser() any {
	return this.User
}

func TestTargetUserSelfOrAdmin(test *testing.T) {
	createTargetType := func(targetUser TargetUser) TargetUserSelfOrAdmin {
		return TargetUserSelfOrAdmin{targetUser}
	}

	createRequest := func(role model.CourseUserRole, target string) *testTargetUserSelfOrAdminRequestType {
		return &testTargetUserSelfOrAdminRequestType{
			APIRequestCourseUserContext: APIRequestCourseUserContext{
				APIRequestUserContext: APIRequestUserContext{
					UserEmail: role.String() + "@test.com",
					UserPass:  util.Sha256HexFromString(role.String()),
				},
				CourseID: "course101",
			},
			User: TargetUserSelfOrAdmin{
				TargetUser{
					Email: target,
				},
			},
		}
	}

	isNonSelfPermError := func(role model.CourseUserRole) bool {
		return role < model.CourseRoleAdmin
	}

	testTargetUser(test, createTargetType, createRequest, isNonSelfPermError)
}

type testTargetUserSelfOrAdminRequestType struct {
	APIRequestCourseUserContext
	MinCourseRoleOther

	User TargetUserSelfOrAdmin
}

func (this *testTargetUserSelfOrAdminRequestType) GetUser() any {
	return this.User
}

type userGetter interface {
	GetUser() any
}

func testTargetUser[T comparable, V userGetter](test *testing.T,
	createTargetType func(TargetUser) T,
	createRequest func(model.CourseUserRole, string) V,
	isNonSelfPermError func(model.CourseUserRole) bool) {
	course := db.MustGetTestCourse()

	users, err := db.GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Failed to get users: '%v'.", err)
	}

	testCases := []struct {
		role      model.CourseUserRole
		target    string
		permError bool
		expected  T
	}{
		// Self.
		{model.CourseRoleStudent, "", false,
			createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
		{model.CourseRoleStudent, "student@test.com", false,
			createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
		{model.CourseRoleGrader, "", false,
			createTargetType(TargetUser{true, "grader@test.com", users["grader@test.com"]})},
		{model.CourseRoleGrader, "grader@test.com", false,
			createTargetType(TargetUser{true, "grader@test.com", users["grader@test.com"]})},

		// Other.
		{model.CourseRoleOther, "student@test.com", isNonSelfPermError(model.CourseRoleOther),
			createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
		{model.CourseRoleStudent, "grader@test.com", isNonSelfPermError(model.CourseRoleStudent),
			createTargetType(TargetUser{true, "grader@test.com", users["grader@test.com"]})},
		{model.CourseRoleGrader, "student@test.com", isNonSelfPermError(model.CourseRoleGrader),
			createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
		{model.CourseRoleAdmin, "student@test.com", isNonSelfPermError(model.CourseRoleAdmin),
			createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},
		{model.CourseRoleOwner, "student@test.com", isNonSelfPermError(model.CourseRoleOwner),
			createTargetType(TargetUser{true, "student@test.com", users["student@test.com"]})},

		// Not found.
		{model.CourseRoleGrader, "ZZZ", isNonSelfPermError(model.CourseRoleGrader),
			createTargetType(TargetUser{false, "ZZZ", nil})},
		{model.CourseRoleAdmin, "ZZZ", isNonSelfPermError(model.CourseRoleAdmin),
			createTargetType(TargetUser{false, "ZZZ", nil})},
	}

	for i, testCase := range testCases {
		request := createRequest(testCase.role, testCase.target)

		apiErr := ValidateAPIRequest(nil, request, "")
		if apiErr != nil {
			if testCase.permError {
				expectedLocator := "-033"
				if expectedLocator != apiErr.Locator {
					test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
						i, expectedLocator, apiErr.Locator)
				}
			} else {
				test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr)
			}

			continue
		}

		if !reflect.DeepEqual(testCase.expected, request.GetUser()) {
			test.Errorf("Case %d: Result not as expected. Expcted '%+v', found '%+v'.",
				i, testCase.expected, request.GetUser())
		}
	}
}

func TestTargetUser(test *testing.T) {
	type requestType struct {
		APIRequestCourseUserContext
		MinCourseRoleOther

		User TargetUser
	}

	course := db.MustGetTestCourse()

	users, err := db.GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Failed to get users: '%v'.", err)
	}

	testCases := []struct {
		role     model.CourseUserRole
		target   string
		expected TargetUser
	}{
		{model.CourseRoleStudent, "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
		{model.CourseRoleGrader, "grader@test.com", TargetUser{true, "grader@test.com", users["grader@test.com"]}},

		{model.CourseRoleStudent, "", TargetUser{}},
		{model.CourseRoleGrader, "", TargetUser{}},

		{model.CourseRoleOther, "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
		{model.CourseRoleStudent, "grader@test.com", TargetUser{true, "grader@test.com", users["grader@test.com"]}},
		{model.CourseRoleGrader, "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
		{model.CourseRoleAdmin, "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},
		{model.CourseRoleOwner, "student@test.com", TargetUser{true, "student@test.com", users["student@test.com"]}},

		// Not found.
		{model.CourseRoleGrader, "ZZZ", TargetUser{false, "ZZZ", nil}},
	}

	for i, testCase := range testCases {
		request := requestType{
			APIRequestCourseUserContext: APIRequestCourseUserContext{
				APIRequestUserContext: APIRequestUserContext{
					UserEmail: testCase.role.String() + "@test.com",
					UserPass:  util.Sha256HexFromString(testCase.role.String()),
				},
				CourseID: "course101",
			},
			User: TargetUser{
				Email: testCase.target,
			},
		}

		apiErr := ValidateAPIRequest(nil, &request, "")
		if apiErr != nil {
			if testCase.target == "" {
				expectedLocator := "-034"
				if expectedLocator != apiErr.Locator {
					test.Errorf("Case %d: Incorrect error returned on empty string. Expcted '%s', found '%s'.",
						i, expectedLocator, apiErr.Locator)
				}
			} else {
				test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr)
			}

			continue
		}

		if !reflect.DeepEqual(testCase.expected, request.User) {
			test.Errorf("Case %d: Result not as expected. Expcted '%+v', found '%+v'.",
				i, testCase.expected, request.User)
		}
	}
}
