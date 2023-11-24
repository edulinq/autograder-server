package core

// How to represent users in API responses.

import (
    "fmt"
    "slices"
    "strings"

    "github.com/eriq-augustine/autograder/model"
)

type UserInfo struct {
    Email string `json:"email"`
    Name string `json:"name"`
    Role model.UserRole `json:"role"`
    LMSID string `json:"lms-id"`
}

type UserInfoWithPass struct {
    UserInfo
    Pass string `json:"pass"`
}

func NewUserInfo(user *model.User) *UserInfo {
    return &UserInfo{
        Email: user.Email,
        Name: user.DisplayName,
        Role: user.Role,
        LMSID: user.LMSID,
    };
}

func (this *UserInfoWithPass) ToUsr() (*model.User, error) {
    if (this.Email == "") {
        return nil, fmt.Errorf("Empty emails are not allowed.")
    }

    user := model.User{
        Email: this.Email,
        Pass: this.Pass,
        DisplayName: this.Name,
        Role: this.Role,
        LMSID: this.LMSID,
    };

    return &user, nil;
}

func NewUserInfos(users []*model.User) []*UserInfo {
    result := make([]*UserInfo, 0, len(users));
    for _, user := range users {
        result = append(result, NewUserInfo(user));
    }

    slices.SortFunc(result, CompareUserInfoPointer);

    return result;
}

func CompareUserInfoPointer(a *UserInfo, b *UserInfo) int {
    if (a == b) {
        return 0;
    }

    if (a == nil) {
        return 1;
    }

    if (b == nil) {
        return -1;
    }

    return strings.Compare(a.Email, b.Email);
}

// Get user info from a generic map (like what an API response would have).
func UserInfoFromMap(data map[string]any) *UserInfo {
    return &UserInfo{
        Email: data["email"].(string),
        Name: data["name"].(string),
        Role: model.GetRole(data["role"].(string)),
        LMSID: data["lms-id"].(string),
    };
}

func CompareUserInfo(a UserInfo, b UserInfo) int {
    return strings.Compare(a.Email, b.Email);
}

// An API-friendly version of model.UserSyncResult.
type SyncUsersInfo struct {
    Add []*UserInfo `json:"add-users"`
    Mod []*UserInfo `json:"mod-users"`
    Del []*UserInfo `json:"del-users"`
    Skip []*UserInfo `json:"skip-users"`
    Unchanged []*UserInfo `json:"unchanged-users"`
}

func NewSyncUsersInfo(syncResult *model.UserSyncResult) *SyncUsersInfo {
    info := SyncUsersInfo{
        Add: NewUserInfos(syncResult.Add),
        Mod: NewUserInfos(syncResult.Mod),
        Del: NewUserInfos(syncResult.Del),
        Skip: NewUserInfos(syncResult.Skip),
        Unchanged: NewUserInfos(syncResult.Unchanged),
    };

    return &info;
}
