package model

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/util"
)

const USERS_FILENAME = "users.json"

// Load users from a file adjacent to the course config (if it exists).
func loadStaticCourseUsers(courseConfigPath string) (map[string]*CourseUser, error) {
	path := filepath.Join(filepath.Dir(courseConfigPath), USERS_FILENAME)
	if !util.PathExists(path) {
		return make(map[string]*CourseUser), nil
	}

	return LoadCourseUsersFile(path)
}

func LoadCourseUsersFile(path string) (map[string]*CourseUser, error) {
	if !util.PathExists(path) {
		return nil, fmt.Errorf("Course users file '%s' does not exist.", path)
	}

	users := make(map[string]*CourseUser)
	err := util.JSONFromFile(path, &users)
	if err != nil {
		return nil, err
	}

	var userErrors error = nil
	for email, user := range users {
		err = user.Validate()
		if err != nil {
			err = fmt.Errorf("Error in user with key '%s': '%w'.", email, err)
			userErrors = errors.Join(userErrors, err)
		}
	}

	if userErrors != nil {
		return nil, fmt.Errorf("Found errors while loading course users file '%s': '%w'.", path, userErrors)
	}

	return users, nil
}

func LoadServerUsersFile(path string) (map[string]*ServerUser, error) {
	if !util.PathExists(path) {
		return nil, fmt.Errorf("Server users file '%s' does not exist.", path)
	}

	users := make(map[string]*ServerUser)
	err := util.JSONFromFile(path, &users)
	if err != nil {
		return nil, err
	}

	var userErrors error = nil
	for email, user := range users {
		err = user.Validate()
		if err != nil {
			err = fmt.Errorf("Error in user with key '%s': '%w'.", email, err)
			userErrors = errors.Join(userErrors, err)
		}
	}

	if userErrors != nil {
		return nil, fmt.Errorf("Found errors while loading server users file '%s': '%w'.", path, userErrors)
	}

	return users, nil
}
