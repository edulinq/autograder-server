package db

import (
	"errors"
	"fmt"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// See Backend.
func GetServerUsers() (map[string]*model.ServerUser, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetServerUsers()
}

func MustGetServerUsers() map[string]*model.ServerUser {
	users, err := GetServerUsers()
	if err != nil {
		log.Fatal("Faled to get server users.", err)
	}

	return users
}

// See Backend.
func GetCourseUsers(course *model.Course) (map[string]*model.CourseUser, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetCourseUsers(course)
}

// See Backend.
func GetServerUser(email string) (*model.ServerUser, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	return backend.GetServerUser(email)
}

func MustGetServerUser(email string) *model.ServerUser {
	user, err := GetServerUser(email)
	if err != nil {
		log.Fatal("Faled to get server user.", err, log.NewUserAttr(email))
	}

	return user
}

// See Backend.
func UpsertUsers(users map[string]*model.ServerUser) error {
	if backend == nil {
		return fmt.Errorf("Database has not been opened.")
	}

	return backend.UpsertUsers(users)
}

// See Backend.
func DeleteUserToken(email string, tokenID string) (bool, error) {
	if backend == nil {
		return false, fmt.Errorf("Database has not been opened.")
	}

	return backend.DeleteUserToken(email, tokenID)
}

// Get a specific course user.
// Returns nil if the user does not exist or is not enrolled in the course.
func GetCourseUser(course *model.Course, email string) (*model.CourseUser, error) {
	if backend == nil {
		return nil, fmt.Errorf("Database has not been opened.")
	}

	serverUser, err := backend.GetServerUser(email)
	if err != nil {
		return nil, err
	}

	if serverUser == nil {
		return nil, nil
	}

	return serverUser.ToCourseUser(course.ID, false)
}

// Convenience function for UpsertUsers() with a single user.
func UpsertUser(user *model.ServerUser) error {
	users := map[string]*model.ServerUser{user.Email: user}
	return UpsertUsers(users)
}

// Convenience function for UpsertUsers() with course users.
func UpsertCourseUsers(course *model.Course, users map[string]*model.CourseUser) error {
	serverUsers := make(map[string]*model.ServerUser, len(users))

	var userErrors error = nil
	for email, user := range users {
		serverUser, err := user.ToServerUser(course.ID)
		if err != nil {
			userErrors = errors.Join(userErrors, fmt.Errorf("Invalid user '%s': '%w'.", email, err))
		} else {
			serverUsers[email] = serverUser
		}
	}

	if userErrors != nil {
		return fmt.Errorf("Found errors when processing users: '%w'.", userErrors)
	}

	return UpsertUsers(serverUsers)
}

// Convenience function for UpsertCourseUsers() with a single user.
func UpsertCourseUser(course *model.Course, user *model.CourseUser) error {
	users := map[string]*model.CourseUser{user.Email: user}
	return UpsertCourseUsers(course, users)
}

// Delete a user from the server.
// Returns a boolean indicating if the user exists.
// If true, then the user exists and was removed.
// If false (and the error is nil), then the user did not exist.
func DeleteUser(email string) (bool, error) {
	if backend == nil {
		return false, fmt.Errorf("Database has not been opened.")
	}

	user, err := GetServerUser(email)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return true, backend.DeleteUser(email)
}

// Remove a user from the course (but leave on the server).
// Returns booleans indicating if the user exists and was enrolled in the course.
func RemoveUserFromCourse(course *model.Course, email string) (bool, bool, error) {
	if backend == nil {
		return false, false, fmt.Errorf("Database has not been opened.")
	}

	user, err := GetServerUser(email)
	if err != nil {
		return false, false, err
	}

	if user == nil {
		return false, false, nil
	}

	_, exists := user.CourseInfo[course.ID]
	if !exists {
		return true, false, nil
	}

	return true, true, backend.RemoveUserFromCourse(course, email)
}

func initializeRootUser() error {
	rootUser := model.ServerUser{
		Email: model.RootUserEmail,
		Role:  model.ServerRoleRoot,
	}

	err := rootUser.Validate()
	if err != nil {
		return fmt.Errorf("Failed to validate the root user: %w", err)
	}

	err = UpsertUser(&rootUser)
	if err != nil {
		return err
	}

	return nil
}
