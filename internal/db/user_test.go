package db

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func (this *DBTests) DBTestUserGetServerUsersBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	expected := mustLoadTestServerUsers()

	users, err := GetServerUsers()
	if err != nil {
		test.Fatalf("Could not get server users: '%v'.", err)
	}

	// Check that root is a server user.
	_, exists := users[model.RootUserEmail]
	if !exists {
		test.Fatalf("Could not find the root user in server users.")
	}

	// Remove root from the server users since we're comparing server users with test server users.
	delete(users, model.RootUserEmail)

	if len(users) == 0 {
		test.Fatalf("Found no server users.")
	}

	if !reflect.DeepEqual(expected, users) {
		test.Fatalf("Server users do not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(users))
	}
}

func (this *DBTests) DBTestUserGetServerUsersEmpty(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	MustClear()

	users, err := GetServerUsers()
	if err != nil {
		test.Fatalf("Could not get server users: '%v'.", err)
	}

	if len(users) == 0 {
		test.Fatalf("Could not find the root user after clearing the database.")
	}

	for _, user := range users {
		if user.Email != model.RootUserEmail {
			test.Fatalf("Found server user '%s' when root should be the only server user.", user.Email)
		}
	}

	if len(users) > 1 {
		test.Fatalf("Found more than one root user in the database.")
	}
}

func (this *DBTests) DBTestUserGetCourseUsersBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()

	serverUsers := mustLoadTestServerUsers()
	expected := convertToCourseUsers(test, course, serverUsers)

	testCourseUsers(test, course, expected)
}

func (this *DBTests) DBTestUserGetCourseUsersEmpty(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetCourse("course-languages")

	users, err := GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Could not get initial course users: '%v'.", err)
	}

	if len(users) == 0 {
		test.Fatalf("Could not find any users when there should be some.")
	}

	// Clear the db (users) and re-add the courses without server-level users..
	MustClear()
	MustAddCourses()

	course = MustGetCourse("course-languages")

	users, err = GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Could not get course users: '%v'.", err)
	}

	if len(users) != 0 {
		test.Fatalf("Found course users when there should have been none: '%s'.", util.MustToJSONIndent(users))
	}
}

func (this *DBTests) DBTestUserGetServerUserBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "course-student@test.edulinq.org"
	expected := mustLoadTestServerUsers()[email]

	user, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, user) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(user))
	}
}

func (this *DBTests) DBTestUserGetServerUserNoTokens(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "course-student@test.edulinq.org"
	expected := mustLoadTestServerUsers()[email]

	user, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, user) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(user))
	}
}

func (this *DBTests) DBTestUserGetServerUserMissing(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "ZZZ"

	user, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got a non-nil server user ('%s') when there should be no user.", email)
	}
}

func (this *DBTests) DBTestUserGetCourseUserBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()
	email := "course-student@test.edulinq.org"

	expected, err := mustLoadTestServerUsers()[email].ToCourseUser(course.ID, false)
	if err != nil {
		test.Fatalf("Could not get expected course user ('%s'): '%v'.", email, err)
	}

	user, err := GetCourseUser(course, email)
	if err != nil {
		test.Fatalf("Could not get course user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil course user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, user) {
		test.Fatalf("Course user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(user))
	}
}

func (this *DBTests) DBTestUserGetCourseUserMissing(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()
	email := "ZZZ"

	user, err := GetCourseUser(course, email)
	if err != nil {
		test.Fatalf("Could not get course user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got a non-nil course user ('%s') when there should be no user.", email)
	}
}

func (this *DBTests) DBTestUserGetCourseUserNotEnrolled(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()
	email := "course-student@test.edulinq.org"

	_, _, err := RemoveUserFromCourse(course, email)
	if err != nil {
		test.Fatalf("Failed to remove user from course: '%v'.", err)
	}

	user, err := GetCourseUser(course, email)
	if err != nil {
		test.Fatalf("Could not get course user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got a non-nil course user ('%s') when there should be no user.", email)
	}
}

func (this *DBTests) DBTestUserUpsertUserInsert(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "new@test.edulinq.org"
	name := "new"

	expected := &model.ServerUser{
		Email: email,
		Name:  &name,
		Role:  model.ServerRoleUser,
	}

	err := expected.Validate()
	if err != nil {
		test.Fatalf("Failed to validated upsert user: '%v'.", err)
	}

	err = UpsertUser(expected)
	if err != nil {
		test.Fatalf("Could not upsert user '%s': '%v'.", email, err)
	}

	newUser, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if newUser == nil {
		test.Fatalf("Got nil (new) server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, newUser) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(newUser))
	}
}

func (this *DBTests) DBTestUserUpsertUserUpdate(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "course-student@test.edulinq.org"
	expected := mustLoadTestServerUsers()[email]

	newExpectedName := "Test Name"
	expected.Name = &newExpectedName

	user, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil server user ('%s').", email)
	}

	newActualName := "Test Name"
	user.Name = &newActualName

	// Remove any additive components.
	user.Tokens = make([]*model.Token, 0)

	err = UpsertUser(user)
	if err != nil {
		test.Fatalf("Could not upsert user '%s': '%v'.", email, err)
	}

	newUser, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if newUser == nil {
		test.Fatalf("Got nil (new) server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, newUser) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(newUser))
	}
}

func (this *DBTests) DBTestUserUpsertUserEmptyUpdate(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "course-student@test.edulinq.org"
	expected := mustLoadTestServerUsers()[email]

	user, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get server user ('%s'): '%v'.", email, err)
	}

	if user == nil {
		test.Fatalf("Got nil server user ('%s').", email)
	}

	// Remove any additive components.
	user.Tokens = make([]*model.Token, 0)

	err = UpsertUser(user)
	if err != nil {
		test.Fatalf("Could not upsert user '%s': '%v'.", email, err)
	}

	newUser, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if newUser == nil {
		test.Fatalf("Got nil (new) server user ('%s').", email)
	}

	if !reflect.DeepEqual(expected, newUser) {
		test.Fatalf("Server user does not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(newUser))
	}
}

func (this *DBTests) DBTestUserDeleteUserBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "course-student@test.edulinq.org"

	exists, err := DeleteUser(email)
	if err != nil {
		test.Fatalf("Could not delete user '%s': '%v'.", email, err)
	}

	if !exists {
		test.Fatalf("Told that user ('%s') did not exist, when it should have.", email)
	}

	user, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got (new) server user ('%s') when it should have been deleted.", email)
	}
}

func (this *DBTests) DBTestUserDeleteUserMissing(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "ZZZ"

	exists, err := DeleteUser(email)
	if err != nil {
		test.Fatalf("Could not delete user '%s': '%v'.", email, err)
	}

	if exists {
		test.Fatalf("Told that user ('%s') exists, when it should not.", email)
	}

	user, err := GetServerUser(email)
	if err != nil {
		test.Fatalf("Could not get (new) server user ('%s'): '%v'.", email, err)
	}

	if user != nil {
		test.Fatalf("Got (new) server user ('%s') when it should have been deleted.", email)
	}
}

func (this *DBTests) DBTestUserRemoveUserFromCourseBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	course := MustGetTestCourse()

	testCases := []struct {
		email    string
		exists   bool
		enrolled bool
	}{
		{"course-student@test.edulinq.org", true, true},
		// Note that we will not reset between test cases.
		{"course-student@test.edulinq.org", true, false},
		{"ZZZ", false, false},
	}

	for i, testCase := range testCases {
		exists, enrolled, err := RemoveUserFromCourse(course, testCase.email)
		if err != nil {
			test.Errorf("Case %d: Failed to remove user ('%s') from course: '%v'.", i, testCase.email, err)
			continue
		}

		if testCase.exists != exists {
			test.Errorf("Case %d: Unexpected exists. Expected: %v, Actual: %v.", i, testCase.exists, exists)
			continue
		}

		if testCase.enrolled != enrolled {
			test.Errorf("Case %d: Unexpected enrolled. Expected: %v, Actual: %v.", i, testCase.enrolled, enrolled)
			continue
		}

		user, err := GetCourseUser(course, testCase.email)
		if err != nil {
			test.Fatalf("Could not get course user ('%s'): '%v'.", testCase.email, err)
		}

		if user != nil {
			test.Fatalf("Got a non-nil course user ('%s') when there should be no user enrolled.", testCase.email)
		}
	}
}

func (this *DBTests) DBTestUserDeleteTokenBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	email := "course-admin@test.edulinq.org"

	// Ensure the token exists.
	user := MustGetServerUser(email)

	if len(user.Tokens) != 1 {
		test.Fatalf("Incorrect number of tokens. Expected: 1, Actual: %d.", len(user.Tokens))
	}

	tokenID := user.Tokens[0].ID

	// Note that we do not reset between cases.
	testCases := []struct {
		email   string
		id      string
		removed bool
	}{
		{email, "ZZZ", false},
		{email, tokenID, true},
		{email, tokenID, false},
		{email, "ZZZ", false},
	}

	for i, testCase := range testCases {
		initialCount := len(MustGetServerUser(email).Tokens)
		expectedCount := initialCount

		removed, err := DeleteUserToken(testCase.email, testCase.id)
		if err != nil {
			test.Fatalf("Case %d: Got an error when removing a token: '%v'.", i, err)
		}

		if testCase.removed != removed {
			test.Fatalf("Case %d: Unexpected removed respose. Expected: %v, Actual: %v.", i, testCase.removed, removed)
		}

		if removed {
			expectedCount--
		}

		newCount := len(MustGetServerUser(email).Tokens)
		if expectedCount != newCount {
			test.Fatalf("Case %d: Unexpected token count. Expected: %v, Actual: %v.", i, expectedCount, newCount)
		}
	}
}

func testCourseUsers(test *testing.T, course *model.Course, expected map[string]*model.CourseUser) {
	users, err := GetCourseUsers(course)
	if err != nil {
		test.Fatalf("Could not get course users: '%v'.", err)
	}

	if len(users) == 0 {
		test.Fatalf("Found no course users.")
	}

	if !reflect.DeepEqual(expected, users) {
		test.Fatalf("Course users do not match. Expected: '%s', Actual: '%s'.",
			util.MustToJSONIndent(expected), util.MustToJSONIndent(users))
	}
}

func mustLoadTestServerUsers() map[string]*model.ServerUser {
	path := filepath.Join(config.GetCourseImportDir(), "testdata", model.USERS_FILENAME)

	users, err := model.LoadServerUsersFile(path)
	if err != nil {
		log.Fatal("Could not open test users file.", err, log.NewAttr("path", path))
	}

	return users
}

func convertToCourseUsers(test *testing.T, course *model.Course, serverUsers map[string]*model.ServerUser) map[string]*model.CourseUser {
	courseUsers := make(map[string]*model.CourseUser, len(serverUsers))
	for email, serverUser := range serverUsers {
		courseUser, err := serverUser.ToCourseUser(course.ID, false)
		if err != nil {
			test.Fatalf("Could not convert server user to course user: '%v'.", err)
		}

		if courseUser != nil {
			courseUsers[email] = courseUser
		}
	}

	return courseUsers
}

func (this *DBTests) DBTestRootUserValidation(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	rootUser, err := GetServerUser(model.RootUserEmail)
	if err != nil {
		test.Fatal("Failed to get the root user.", err)
	}

	if rootUser == nil {
		test.Fatal("Root user not found.")
	}

	err = rootUser.Validate()
	if err != nil {
		test.Fatal("Root user validation failed.", err)
	}
}
