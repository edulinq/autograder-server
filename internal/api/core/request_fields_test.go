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

// Test:
// CourseUsers,
// TargetServerUser, TargetServerUserSelfOrAdmin,
// TargetCourseUser, TargetCourseUserSelfOrGrader, and TargetCourseUserSelfOrAdmin.
// No embeded course context.
func TestBadUsersFieldNoContext(test *testing.T) {
	testCases := []struct {
		locator string
		request any
	}{
		{"-025", &struct{ Users CourseUsers }{}},
		{"-042", &struct{ User TargetServerUser }{}},
		{"-042", &struct{ User TargetServerUserSelfOrAdmin }{}},
		{"-025", &struct{ User TargetCourseUser }{}},
		{"-025", &struct{ User TargetCourseUserSelfOrGrader }{}},
		{"-025", &struct{ User TargetCourseUserSelfOrAdmin }{}},
	}

	for i, testCase := range testCases {
		apiErr := checkRequestSpecialFields(nil, testCase.request, "")
		if apiErr == nil {
			test.Fatalf("Case %d: Struct with no course context does not return an error: '%+v'.",
				i, testCase.request)
		}

		if testCase.locator != apiErr.Locator {
			test.Fatalf("Case %d: Struct with no course context does not return an error with locator '%s', found '%s': '%+v.",
				i, testCase.locator, apiErr.Locator, testCase.request)
		}
	}
}

// Test:
// CourseUsers,
// TargetServerUser, TargetServerUserSelfOrAdmin,
// TargetCourseUser, TargetCourseUserSelfOrGrader, and TargetCourseUserSelfOrAdmin.
// Users are not exported.
func TestBadUsersFieldNotExported(test *testing.T) {
	testCases := []struct {
		locator string
		request any
	}{
		{
			"-026", &struct {
				APIRequestCourseUserContext
				MinCourseRoleStudent
				users CourseUsers
			}{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "course-student@test.edulinq.org",
						UserPass:  studentPass,
					},
					CourseID: "course101",
				},
			},
		},
		{
			"-043", &struct {
				APIRequestUserContext
				targetServerUser TargetServerUser
			}{
				APIRequestUserContext: APIRequestUserContext{
					UserEmail: "course-student@test.edulinq.org",
					UserPass:  studentPass,
				},
			},
		},
		{
			"-043", &struct {
				APIRequestUserContext
				targetServerUser TargetServerUserSelfOrAdmin
			}{
				APIRequestUserContext: APIRequestUserContext{
					UserEmail: "course-student@test.edulinq.org",
					UserPass:  studentPass,
				},
			},
		},
		{
			"-026", &struct {
				APIRequestCourseUserContext
				MinCourseRoleStudent
				targetCourseUser TargetCourseUser
			}{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "course-student@test.edulinq.org",
						UserPass:  studentPass,
					},
					CourseID: "course101",
				},
			},
		},
		{
			"-026", &struct {
				APIRequestCourseUserContext
				MinCourseRoleStudent
				targetCourseUser TargetCourseUserSelfOrGrader
			}{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "course-student@test.edulinq.org",
						UserPass:  studentPass,
					},
					CourseID: "course101",
				},
			},
		},
		{
			"-026", &struct {
				APIRequestCourseUserContext
				MinCourseRoleStudent
				targetCourseUser TargetCourseUserSelfOrAdmin
			}{
				APIRequestCourseUserContext: APIRequestCourseUserContext{
					APIRequestUserContext: APIRequestUserContext{
						UserEmail: "course-student@test.edulinq.org",
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

		if testCase.locator != apiErr.Locator {
			test.Fatalf("Case %d: Struct with non-exported field does not return an error with locator '%s', found '%s': '%v.",
				i, testCase.locator, apiErr.Locator, apiErr)
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

	response := SendTestAPIRequestFull(test, endpoint, nil, paths, "course-admin@test.edulinq.org")
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

	response := SendTestAPIRequestFull(test, endpoint, nil, paths, "course-admin@test.edulinq.org")
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

	response := SendTestAPIRequestFull(test, endpoint, nil, paths, "course-admin@test.edulinq.org")
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

	response := SendTestAPIRequestFull(test, endpoint, nil, paths, "course-admin@test.edulinq.org")
	if response.Success {
		test.Fatalf("Request did not generate an error: '%v'.", response)
	}

	expectedLocator := "-036"
	if response.Locator != expectedLocator {
		test.Fatalf("Error does not have the correct locator. Expected '%s', found '%s'.",
			expectedLocator, response.Locator)
	}
}

func TestTargetCourseUserJSON(test *testing.T) {
	createTargetType := func(targetCourseUser TargetCourseUser) TargetCourseUser {
		return targetCourseUser
	}

	testTargetCourseUserJSON(test, createTargetType)
}

func TestTargetCourseUserSelfOrGraderJSON(test *testing.T) {
	createTargetType := func(targetCourseUser TargetCourseUser) TargetCourseUserSelfOrGrader {
		return TargetCourseUserSelfOrGrader{targetCourseUser}
	}

	testTargetCourseUserJSON(test, createTargetType)
}

func TestTargetCourseUserSelfOrAdminJSON(test *testing.T) {
	createTargetType := func(targetCourseUser TargetCourseUser) TargetCourseUserSelfOrAdmin {
		return TargetCourseUserSelfOrAdmin{targetCourseUser}
	}

	testTargetCourseUserJSON(test, createTargetType)
}

func testTargetCourseUserJSON[T comparable](test *testing.T, createTargetType func(TargetCourseUser) T) {
	testCases := []struct {
		in       string
		expected T
	}{
		{`""`, createTargetType(TargetCourseUser{false, "", nil})},
		{`"a"`, createTargetType(TargetCourseUser{false, "a", nil})},
		{`"course-student@test.edulinq.org"`, createTargetType(TargetCourseUser{false, "course-student@test.edulinq.org", nil})},
		{`"a\"b\"c"`, createTargetType(TargetCourseUser{false, `a"b"c`, nil})},
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

func TestTargetServerUserJSON(test *testing.T) {
	createTargetType := func(targetServerUser TargetServerUser) TargetServerUser {
		return targetServerUser
	}

	testTargetServerUserJSON(test, createTargetType)
}

func TestTargetServerUserSelfOrAdminJSON(test *testing.T) {
	createTargetType := func(targetServerUser TargetServerUser) TargetServerUserSelfOrAdmin {
		return TargetServerUserSelfOrAdmin{targetServerUser}
	}

	testTargetServerUserJSON(test, createTargetType)
}

func testTargetServerUserJSON[T comparable](test *testing.T, createTargetType func(TargetServerUser) T) {
	testCases := []struct {
		in       string
		expected T
	}{
		{`""`, createTargetType(TargetServerUser{false, "", nil})},
		{`"a"`, createTargetType(TargetServerUser{false, "a", nil})},
		{`"course-student@test.edulinq.org"`, createTargetType(TargetServerUser{false, "course-student@test.edulinq.org", nil})},
		{`"a\"b\"c"`, createTargetType(TargetServerUser{false, `a"b"c`, nil})},
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

func TestTargetCourseUserSelfOrGrader(test *testing.T) {
	createTargetType := func(targetCourseUser TargetCourseUser) TargetCourseUserSelfOrGrader {
		return TargetCourseUserSelfOrGrader{targetCourseUser}
	}

	createRequest := func(role model.CourseUserRole, target string) *testTargetCourseUserSelfOrGraderRequestType {
		return &testTargetCourseUserSelfOrGraderRequestType{
			APIRequestCourseUserContext: APIRequestCourseUserContext{
				APIRequestUserContext: APIRequestUserContext{
					UserEmail: role.String() + "@test.edulinq.org",
					UserPass:  util.Sha256HexFromString(role.String()),
				},
				CourseID: "course101",
			},
			User: TargetCourseUserSelfOrGrader{
				TargetCourseUser{
					Email: target,
				},
			},
		}
	}

	isNonSelfPermError := func(role model.CourseUserRole) bool {
		return role < model.CourseRoleGrader
	}

	testTargetCourseUser(test, createTargetType, createRequest, isNonSelfPermError)
}

type testTargetCourseUserSelfOrGraderRequestType struct {
	APIRequestCourseUserContext
	MinCourseRoleOther

	User TargetCourseUserSelfOrGrader
}

func (this *testTargetCourseUserSelfOrGraderRequestType) GetUser() any {
	return this.User
}

func TestTargetCourseUserSelfOrAdmin(test *testing.T) {
	createTargetType := func(targetCourseUser TargetCourseUser) TargetCourseUserSelfOrAdmin {
		return TargetCourseUserSelfOrAdmin{targetCourseUser}
	}

	createRequest := func(role model.CourseUserRole, target string) *testTargetCourseUserSelfOrAdminRequestType {
		return &testTargetCourseUserSelfOrAdminRequestType{
			APIRequestCourseUserContext: APIRequestCourseUserContext{
				APIRequestUserContext: APIRequestUserContext{
					UserEmail: role.String() + "@test.edulinq.org",
					UserPass:  util.Sha256HexFromString(role.String()),
				},
				CourseID: "course101",
			},
			User: TargetCourseUserSelfOrAdmin{
				TargetCourseUser{
					Email: target,
				},
			},
		}
	}

	isNonSelfPermError := func(role model.CourseUserRole) bool {
		return role < model.CourseRoleAdmin
	}

	testTargetCourseUser(test, createTargetType, createRequest, isNonSelfPermError)
}

type testTargetCourseUserSelfOrAdminRequestType struct {
	APIRequestCourseUserContext
	MinCourseRoleOther

	User TargetCourseUserSelfOrAdmin
}

func (this *testTargetCourseUserSelfOrAdminRequestType) GetUser() any {
	return this.User
}

type userGetter interface {
	GetUser() any
}

func testTargetCourseUser[T comparable, V userGetter](test *testing.T,
	createTargetType func(TargetCourseUser) T,
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
			createTargetType(TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]})},
		{model.CourseRoleStudent, "course-student@test.edulinq.org", false,
			createTargetType(TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]})},
		{model.CourseRoleGrader, "", false,
			createTargetType(TargetCourseUser{true, "course-grader@test.edulinq.org", users["course-grader@test.edulinq.org"]})},
		{model.CourseRoleGrader, "course-grader@test.edulinq.org", false,
			createTargetType(TargetCourseUser{true, "course-grader@test.edulinq.org", users["course-grader@test.edulinq.org"]})},

		// Other.
		{model.CourseRoleOther, "course-student@test.edulinq.org", isNonSelfPermError(model.CourseRoleOther),
			createTargetType(TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]})},
		{model.CourseRoleStudent, "course-grader@test.edulinq.org", isNonSelfPermError(model.CourseRoleStudent),
			createTargetType(TargetCourseUser{true, "course-grader@test.edulinq.org", users["course-grader@test.edulinq.org"]})},
		{model.CourseRoleGrader, "course-student@test.edulinq.org", isNonSelfPermError(model.CourseRoleGrader),
			createTargetType(TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]})},
		{model.CourseRoleAdmin, "course-student@test.edulinq.org", isNonSelfPermError(model.CourseRoleAdmin),
			createTargetType(TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]})},
		{model.CourseRoleOwner, "course-student@test.edulinq.org", isNonSelfPermError(model.CourseRoleOwner),
			createTargetType(TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]})},

		// Not found.
		{model.CourseRoleGrader, "ZZZ", isNonSelfPermError(model.CourseRoleGrader),
			createTargetType(TargetCourseUser{false, "ZZZ", nil})},
		{model.CourseRoleAdmin, "ZZZ", isNonSelfPermError(model.CourseRoleAdmin),
			createTargetType(TargetCourseUser{false, "ZZZ", nil})},
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

		if testCase.permError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, request.GetUser()) {
			test.Errorf("Case %d: Result not as expected. Expcted '%+v', found '%+v'.",
				i, testCase.expected, request.GetUser())
		}
	}
}

func TestTargetCourseUser(test *testing.T) {
	type requestType struct {
		APIRequestCourseUserContext
		MinCourseRoleOther

		User TargetCourseUser
	}

	course := db.MustGetTestCourse()

	users, err := db.GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Failed to get users: '%v'.", err)
	}

	testCases := []struct {
		role     model.CourseUserRole
		target   string
		expected TargetCourseUser
	}{
		{model.CourseRoleStudent, "course-student@test.edulinq.org", TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]}},
		{model.CourseRoleGrader, "course-grader@test.edulinq.org", TargetCourseUser{true, "course-grader@test.edulinq.org", users["course-grader@test.edulinq.org"]}},

		{model.CourseRoleStudent, "", TargetCourseUser{}},
		{model.CourseRoleGrader, "", TargetCourseUser{}},

		{model.CourseRoleOther, "course-student@test.edulinq.org", TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]}},
		{model.CourseRoleStudent, "course-grader@test.edulinq.org", TargetCourseUser{true, "course-grader@test.edulinq.org", users["course-grader@test.edulinq.org"]}},
		{model.CourseRoleGrader, "course-student@test.edulinq.org", TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]}},
		{model.CourseRoleAdmin, "course-student@test.edulinq.org", TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]}},
		{model.CourseRoleOwner, "course-student@test.edulinq.org", TargetCourseUser{true, "course-student@test.edulinq.org", users["course-student@test.edulinq.org"]}},

		// Not found.
		{model.CourseRoleGrader, "ZZZ", TargetCourseUser{false, "ZZZ", nil}},
	}

	for i, testCase := range testCases {
		request := requestType{
			APIRequestCourseUserContext: APIRequestCourseUserContext{
				APIRequestUserContext: APIRequestUserContext{
					UserEmail: testCase.role.String() + "@test.edulinq.org",
					UserPass:  util.Sha256HexFromString(testCase.role.String()),
				},
				CourseID: "course101",
			},
			User: TargetCourseUser{
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

func TestTargetServerUser(test *testing.T) {
	type requestType struct {
		APIRequestUserContext

		User TargetServerUser
	}

	user := db.MustGetServerUser("course-student@test.edulinq.org", true)

	testCases := []struct {
		target string
		found  bool
	}{
		{"", false},
		{"course-student@test.edulinq.org", true},
		{"ZZZ", false},
	}

	for i, testCase := range testCases {
		request := requestType{
			APIRequestUserContext: APIRequestUserContext{
				UserEmail: "course-admin@test.edulinq.org",
				UserPass:  util.Sha256HexFromString("admin"),
			},
			User: TargetServerUser{
				Email: testCase.target,
			},
		}

		apiErr := ValidateAPIRequest(nil, &request, "")
		if apiErr != nil {
			if testCase.target == "" {
				expectedLocator := "-044"
				if expectedLocator != apiErr.Locator {
					test.Errorf("Case %d: Incorrect error returned on empty string. Expcted '%s', found '%s'.",
						i, expectedLocator, apiErr.Locator)
				}
			} else {
				test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr)
			}

			continue
		}

		if testCase.found != request.User.Found {
			test.Errorf("Case %d: Found result not as expected. Expcted '%v', found '%v'.",
				i, testCase.found, request.User.Found)
			continue
		}

		if testCase.found {
			if !reflect.DeepEqual(user, request.User.User) {
				test.Errorf("Case %d: Result not as expected. Expcted '%s', found '%s'.",
					i, util.MustToJSONIndent(user), util.MustToJSONIndent(request.User.User))
				continue
			}
		} else {
			if request.User.User != nil {
				test.Errorf("Case %d: Did not get a nil user on !found. Got: '%s'.",
					i, util.MustToJSONIndent(request.User))
				continue
			}
		}
	}
}

func TestTargetServerUserSelfOrAdmin(test *testing.T) {
	users := db.MustGetServerUsers()

	type requestType struct {
		APIRequestUserContext
		MinServerRoleUser

		User TargetServerUserSelfOrAdmin
	}

	testCases := []struct {
		contextUser *model.ServerUser
		target      string
		permError   bool
		expected    *model.ServerUser
	}{
		// Self, empty.
		{users["server-user@test.edulinq.org"], "", false, users["server-user@test.edulinq.org"]},
		{users["server-creator@test.edulinq.org"], "", false, users["server-creator@test.edulinq.org"]},
		{users["server-admin@test.edulinq.org"], "", false, users["server-admin@test.edulinq.org"]},
		{users["server-owner@test.edulinq.org"], "", false, users["server-owner@test.edulinq.org"]},

		// Self, email.
		{users["server-user@test.edulinq.org"], "server-user@test.edulinq.org", false, users["server-user@test.edulinq.org"]},
		{users["server-creator@test.edulinq.org"], "server-creator@test.edulinq.org", false, users["server-creator@test.edulinq.org"]},
		{users["server-admin@test.edulinq.org"], "server-admin@test.edulinq.org", false, users["server-admin@test.edulinq.org"]},
		{users["server-owner@test.edulinq.org"], "server-owner@test.edulinq.org", false, users["server-owner@test.edulinq.org"]},

		// Other, bad permissions.
		{users["server-creator@test.edulinq.org"], "server-user@test.edulinq.org", true, nil},
		{users["server-creator@test.edulinq.org"], "server-admin@test.edulinq.org", true, nil},
		{users["server-creator@test.edulinq.org"], "server-owner@test.edulinq.org", true, nil},

		// Other, good permissions.
		{users["server-admin@test.edulinq.org"], "server-user@test.edulinq.org", false, users["server-user@test.edulinq.org"]},
		{users["server-admin@test.edulinq.org"], "server-creator@test.edulinq.org", false, users["server-creator@test.edulinq.org"]},
		{users["server-admin@test.edulinq.org"], "server-owner@test.edulinq.org", false, users["server-owner@test.edulinq.org"]},

		// Not found.
		{users["server-creator@test.edulinq.org"], "ZZZ", true, nil},
		{users["server-admin@test.edulinq.org"], "ZZZ", false, nil},
	}

	for i, testCase := range testCases {
		request := &requestType{
			APIRequestUserContext: APIRequestUserContext{
				UserEmail: testCase.contextUser.Email,
				UserPass:  util.Sha256HexFromString(*testCase.contextUser.Name),
			},
			User: TargetServerUserSelfOrAdmin{
				TargetServerUser{
					Email: testCase.target,
				},
			},
		}

		apiErr := ValidateAPIRequest(nil, request, "")
		if apiErr != nil {
			if testCase.permError {
				expectedLocator := "-046"
				if expectedLocator != apiErr.Locator {
					test.Errorf("Case %d: Incorrect error returned on permissions error. Expcted '%s', found '%s'.",
						i, expectedLocator, apiErr.Locator)
				}
			} else {
				test.Errorf("Case %d: Failed to validate request: '%v'.", i, apiErr)
			}

			continue
		}

		if testCase.permError {
			test.Errorf("Case %d: Did not get an expected permissions error.", i)
			continue
		}

		// Not found case.
		if testCase.expected == nil {
			if request.User.Found {
				test.Errorf("Case %d: User found when it was not expected.", i)
			}

			continue
		}

		if !request.User.Found {
			test.Errorf("Case %d: User not found when it was expected.", i)
		}

		if !reflect.DeepEqual(testCase.expected, request.User.User) {
			test.Errorf("Case %d: Result not as expected. Expcted '%+v', found '%+v'.",
				i, testCase.expected, request.User.User)
		}
	}
}
