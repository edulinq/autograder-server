package core

// How to represent users in API responses.

import (
    "strings"

    "github.com/eriq-augustine/autograder/usr"
)

type UserInfo struct {
    Email string `json:"email"`
    Name string `json:"name"`
    Role usr.UserRole `json:"role"`
    LMSID string `json:"lms-id"`
}

func NewUserInfo(user *usr.User) *UserInfo {
    return &UserInfo{
        Email: user.Email,
        Name: user.DisplayName,
        Role: user.Role,
        LMSID: user.LMSID,
    };
}

func NewUserInfos(users []*usr.User) []*UserInfo {
    result := make([]*UserInfo, 0, len(users));
    for _, user := range users {
        result = append(result, NewUserInfo(user));
    }
    return result;
}

// Get user info from a generic map (like what an API response would have).
func UserInfoFromMap(data map[string]any) *UserInfo {
    return &UserInfo{
        Email: data["email"].(string),
        Name: data["name"].(string),
        Role: usr.GetRole(data["role"].(string)),
        LMSID: data["lms-id"].(string),
    };
}

func CompareUserInfo(a UserInfo, b UserInfo) int {
    return strings.Compare(a.Email, b.Email);
}
