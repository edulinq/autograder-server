package test

// A fake LMS adapter for testing that reads config from a test course directory.

import (
    "fmt"

    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/usr"
)

func (this *TestLMSAdapter) FetchUsers() ([]*lms.User, error) {
    localUsers, err := this.SourceCourse.GetUsers();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to get local users: '%w'.", err);
    }

    users := make([]*lms.User, 0, len(localUsers));
    for _, localUser := range localUsers {
        users = append(users, UserFromAGUser(localUser));
    }

    if (this.UsersModifier != nil) {
        return this.UsersModifier(users), nil;
    }

    return users, nil;
}

func (this *TestLMSAdapter) SetUsersModifier(modifier FetchUsersModifier) {
    this.UsersModifier = modifier;
}

func (this *TestLMSAdapter) ClearUsersModifier() {
    this.UsersModifier = nil;
}

func (this *TestLMSAdapter) FetchUser(email string) (*lms.User, error) {
    users, err := this.SourceCourse.GetUsers();
    if (err != nil) {
        return nil, fmt.Errorf("Failed to get users: '%w'.", err);
    }

    user := users[email];
    if (user == nil) {
        return nil, nil;
    }

    return UserFromAGUser(user), nil;
}

func UserFromAGUser(user *usr.User) *lms.User {
    return &lms.User{
        ID: "lms-" + user.Email,
        Name: user.DisplayName,
        Email: user.Email,
        Role: user.Role,
    };
}
