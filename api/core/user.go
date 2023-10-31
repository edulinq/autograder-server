package core

// How to represent users in API responses.

import (
    "fmt"
    "strings"

    "github.com/eriq-augustine/autograder/usr"
)

type UserInfo struct {
    Email string `json:"email"`
    Name string `json:"name"`
    Role usr.UserRole `json:"role"`
    LMSID string `json:"lms-id"`
}

type UserInfoWithPass struct {
    UserInfo
    Pass string `json:"pass"`
}

func NewUserInfo(user *usr.User) *UserInfo {
    return &UserInfo{
        Email: user.Email,
        Name: user.DisplayName,
        Role: user.Role,
        LMSID: user.LMSID,
    };
}

func (this *UserInfoWithPass) ToUsr() (*usr.User, error) {
    if (this.Email == "") {
        return nil, fmt.Errorf("Empty emails are not allowed.")
    }

    user := usr.User{
        Email: this.Email,
        Pass: this.Pass,
        DisplayName: this.Name,
        Role: this.Role,
        LMSID: this.LMSID,
    };

    return &user, nil;
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

// An API-friendly version of usr.UserSyncResult.
type SyncUsersInfo struct {
    Add []*UserInfo `json:"add-users"`
    Mod []*UserInfo `json:"mod-users"`
    Del []*UserInfo `json:"del-users"`
    Skip []*UserInfo `json:"skip-users"`
}

func NewSyncUsersInfo(syncResult *usr.UserSyncResult) *SyncUsersInfo {
    info := SyncUsersInfo{
        Add: NewUserInfos(syncResult.Add),
        Mod: NewUserInfos(syncResult.Mod),
        Del: NewUserInfos(syncResult.Del),
        Skip: NewUserInfos(syncResult.Skip),
    };

    return &info;
}
