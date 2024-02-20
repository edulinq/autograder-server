package disk

import (
	"fmt"
	"path/filepath"

	"github.com/eriq-augustine/autograder/model"
	"github.com/eriq-augustine/autograder/util"
)

func (this *backend) GetUsers(course *model.Course) (map[string]*model.User, error) {
	return this.getUsersLock(course, true)
}

func (this *backend) getUsersLock(course *model.Course, acquireLock bool) (map[string]*model.User, error) {
	if acquireLock {
		this.lock.RLock()
		defer this.lock.RUnlock()
	}

	users := make(map[string]*model.User)

	path := this.getUsersPath(course)
	if !util.PathExists(path) {
		return users, nil
	}

	err := util.JSONFromFile(path, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (this *backend) GetUser(course *model.Course, email string) (*model.User, error) {
	users, err := this.GetUsers(course)
	if err != nil {
		return nil, fmt.Errorf("Failed to get users when searching for '%s': '%w'.", email, err)
	}

	return users[email], nil
}

func (this *backend) SaveUsers(course *model.Course, users map[string]*model.User) error {
	return this.saveUsersLock(course, users, true)
}

func (this *backend) saveUsersLock(course *model.Course, newUsers map[string]*model.User, acquireLock bool) error {
	if acquireLock {
		this.lock.Lock()
		defer this.lock.Unlock()
	}

	users, err := this.getUsersLock(course, false)
	if err != nil {
		return fmt.Errorf("Failed to get users to merge before saving: '%w'.", err)
	}

	for key, value := range newUsers {
		users[key] = value
	}

	err = util.ToJSONFileIndent(users, this.getUsersPath(course))
	if err != nil {
		return fmt.Errorf("Unable to save user's file: '%w'.", err)
	}

	return nil
}

func (this *backend) RemoveUser(course *model.Course, email string) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	users, err := this.getUsersLock(course, false)
	if err != nil {
		return fmt.Errorf("Failed to get users when removing for '%s': '%w'.", email, err)
	}

	_, ok := users[email]
	if !ok {
		return nil
	}

	delete(users, email)

	err = util.ToJSONFileIndent(users, this.getUsersPath(course))
	if err != nil {
		return fmt.Errorf("Unable to save user's file: '%w'.", err)
	}

	return nil
}

func (this *backend) getUsersPath(course *model.Course) string {
	return filepath.Join(this.getCourseDir(course), model.USERS_FILENAME)
}
