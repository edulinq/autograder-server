package disk

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/edulinq/autograder/internal/model"
	"github.com/edulinq/autograder/internal/util"
)

func (this *backend) GetServerUsers() (map[string]*model.ServerUser, error) {
	return this.getServerUsersLock(true)
}

func (this *backend) GetCourseUsers(course *model.Course) (map[string]*model.CourseUser, error) {
	users, err := this.getServerUsersLock(true)
	if err != nil {
		return nil, err
	}

	courseUsers := make(map[string]*model.CourseUser)
	for email, user := range users {
		courseUser, err := user.GetCourseUser(course.ID)
		if err != nil {
			return nil, fmt.Errorf("Invalid user '%s': '%w'.", email, err)
		}

		if courseUser != nil {
			courseUsers[courseUser.Email] = courseUser
		}
	}

	return courseUsers, nil
}

func (this *backend) GetServerUser(email string, includeTokens bool) (*model.ServerUser, error) {
	users, err := this.getServerUsersLock(true)
	if err != nil {
		return nil, err
	}

	user, exists := users[email]
	if !exists {
		return nil, nil
	}

	if !includeTokens {
		user.Tokens = nil
	}

	return user, nil
}

func (this *backend) UpsertUsers(users map[string]*model.ServerUser) error {
	return this.upsertUsersLock(users, true)
}

func (this *backend) DeleteUser(email string) error {
	this.userLock.Lock()
	defer this.userLock.Unlock()

	users, err := this.getServerUsersLock(false)
	if err != nil {
		return fmt.Errorf("Failed to get users when deleting user '%s': '%w'.", email, err)
	}

	_, ok := users[email]
	if !ok {
		return nil
	}

	delete(users, email)

	err = util.ToJSONFileIndent(users, this.getServerUsersPath())
	if err != nil {
		return fmt.Errorf("Unable to save user's file: '%w'.", err)
	}

	return nil
}

func (this *backend) RemoveUserFromCourse(course *model.Course, email string) error {
	this.userLock.Lock()
	defer this.userLock.Unlock()

	users, err := this.getServerUsersLock(false)
	if err != nil {
		return fmt.Errorf("Failed to get users when deleting user '%s': '%w'.", email, err)
	}

	user, ok := users[email]
	if !ok {
		return nil
	}

	_, enrolled := user.Roles[course.ID]
	if !enrolled {
		return nil
	}

	delete(user.Roles, course.ID)

	err = util.ToJSONFileIndent(users, this.getServerUsersPath())
	if err != nil {
		return fmt.Errorf("Unable to save user's file: '%w'.", err)
	}

	return nil
}

func (this *backend) getServerUsersLock(acquireLock bool) (map[string]*model.ServerUser, error) {
	if acquireLock {
		this.userLock.RLock()
		defer this.userLock.RUnlock()
	}

	users := make(map[string]*model.ServerUser)

	path := this.getServerUsersPath()
	if !util.PathExists(path) {
		return users, nil
	}

	err := util.JSONFromFile(path, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (this *backend) upsertUsersLock(upsertUsers map[string]*model.ServerUser, acquireLock bool) error {
	if acquireLock {
		this.userLock.Lock()
		defer this.userLock.Unlock()
	}

	users, err := this.getServerUsersLock(false)
	if err != nil {
		return fmt.Errorf("Failed to get users to merge before saving: '%w'.", err)
	}

	for email, upsertUser := range upsertUsers {
		if upsertUser == nil {
			continue
		}

		oldUser, exists := users[email]
		if exists {
			err = oldUser.Merge(upsertUser)
			if err != nil {
				return fmt.Errorf("User '%s' could not be merged with existing user: '%w'.", email, err)
			}
		} else {
			users[email] = upsertUser
		}
	}

	err = util.ToJSONFileIndent(users, this.getServerUsersPath())
	if err != nil {
		return fmt.Errorf("Unable to save user's file: '%w'.", err)
	}

	return nil
}

func (this *backend) getServerUsersPath() string {
	return filepath.Join(this.baseDir, model.USERS_FILENAME)
}

func convertCourseUsers(courseUsers map[string]*model.CourseUser, course *model.Course) (map[string]*model.ServerUser, error) {
	serverUsers := make(map[string]*model.ServerUser, len(courseUsers))

	var userErrors error = nil
	for email, courseUser := range courseUsers {
		serverUser, err := courseUser.GetServerUser(course.ID)
		if err != nil {
			userErrors = errors.Join(userErrors, fmt.Errorf("Error with user '%s': '%w'.", email, err))
		} else {
			serverUsers[email] = serverUser
		}
	}

	if userErrors != nil {
		return nil, userErrors
	}

	return serverUsers, nil
}
