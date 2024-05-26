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

// TEST
/*
	GetServerUsers() (map[string]*model.ServerUser, error)
	GetCourseUsers(course *model.Course) (map[string]*model.CourseUser, error)
	GetServerUser(email string, includeTokens bool) (*model.ServerUser, error)
func GetCourseUser(course *model.Course, email string) (*model.CourseUser, error) {
func UpsertUser(user *model.ServerUser) error {
func UpsertCourseUsers(course *model.Course, users map[string]*model.CourseUser) error {
func UpsertCourseUser(course *model.Course, user *model.CourseUser) error {
func DeleteUser(email string) (bool, error) {
func RemoveUserFromCourse(course *model.Course, email string) (bool, bool, error) {
	UpsertUsers(users map[string]*model.ServerUser) error
	DeleteUser(email string) error
	RemoveUserFromCourse(course *model.Course, email string) error
*/

func (this *DBTests) DBTestGetServerUsersBase(test *testing.T) {
	defer ResetForTesting()
	ResetForTesting()

	// TEST
	// course := MustGetTestCourse()

	expectedServerUsers := mustLoadTestServerUsers()

	users, err := GetServerUsers()
	if err != nil {
		test.Fatalf("Could not get server users: '%v'.", err)
	}

	if len(users) == 0 {
		test.Fatalf("Found no server users.")
	}

	if !reflect.DeepEqual(expectedServerUsers, users) {
		test.Fatalf("Server users do not match. Expected: '%s', actual: '%s'.",
			util.MustToJSONIndent(expectedServerUsers), util.MustToJSONIndent(users))
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
