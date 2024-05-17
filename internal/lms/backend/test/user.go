package test

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/model"
)

func (this *TestLMSBackend) FetchUsers() ([]*lmstypes.User, error) {
	localUsers, err := db.GetUsersFromID(this.CourseID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get local users: '%w'.", err)
	}

	users := make([]*lmstypes.User, 0, len(localUsers))
	for _, localUser := range localUsers {
		users = append(users, UserFromAGUser(localUser))
	}

	if usersModifier != nil {
		return usersModifier(users), nil
	}

	return users, nil
}

func (this *TestLMSBackend) FetchUser(email string) (*lmstypes.User, error) {
	users, err := db.GetUsersFromID(this.CourseID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get users: '%w'.", err)
	}

	user := users[email]
	if user == nil {
		return nil, nil
	}

	return UserFromAGUser(user), nil
}

func UserFromAGUser(user *model.User) *lmstypes.User {
	return &lmstypes.User{
		ID:    "lms-" + user.Email,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}
}
