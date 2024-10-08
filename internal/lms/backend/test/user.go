package test

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/lms/lmstypes"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

func (this *TestLMSBackend) FetchUsers() ([]*lmstypes.User, error) {
	course, err := db.GetCourse(this.CourseID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get course for local users: '%w'.", err)
	}

	if course == nil {
		log.Warn("Could not find course for LMS adapter.", log.NewCourseAttr(this.CourseID))
		return make([]*lmstypes.User, 0), nil
	}

	courseUsers, err := db.GetCourseUsers(course)
	if err != nil {
		return nil, fmt.Errorf("Failed to get local users: '%w'.", err)
	}

	users := make([]*lmstypes.User, 0, len(courseUsers))
	for _, courseUser := range courseUsers {
		users = append(users, UserFromCourseUser(courseUser))
	}

	if usersModifier != nil {
		return usersModifier(users), nil
	}

	return users, nil
}

func (this *TestLMSBackend) FetchUser(email string) (*lmstypes.User, error) {
	course, err := db.GetCourse(this.CourseID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get course for local users: '%w'.", err)
	}

	users, err := db.GetCourseUsers(course)
	if err != nil {
		return nil, fmt.Errorf("Failed to get users: '%w'.", err)
	}

	user := users[email]
	if user == nil {
		return nil, nil
	}

	return UserFromCourseUser(user), nil
}

func UserFromCourseUser(user *model.CourseUser) *lmstypes.User {
	return &lmstypes.User{
		ID:    "lms-" + user.Email,
		Name:  user.GetName(false),
		Email: user.Email,
		Role:  user.Role,
	}
}
