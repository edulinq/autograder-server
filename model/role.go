package model

import (
    "bytes"
    "encoding/json"
    "fmt"
    "strings"
)

type UserRole int;

const (
    RoleUnknown UserRole = 0
    RoleOther            = 10
    RoleStudent          = 20
    RoleGrader           = 30
    RoleAdmin            = 40
    RoleOwner            = 50
)

func (this UserRole) String() string {
    return roleToString[this]
}

var roleToString = map[UserRole]string{
    RoleOwner:   "owner",
    RoleAdmin:   "admin",
    RoleGrader:  "grader",
    RoleStudent: "student",
    RoleOther:   "other",
    RoleUnknown: "unknown",
}

var stringToRole = map[string]UserRole{
    "owner":   RoleOwner,
    "admin":   RoleAdmin,
    "grader":  RoleGrader,
    "student": RoleStudent,
    "other":   RoleOther,
    "unknown": RoleUnknown,
}

func GetRole(text string) UserRole {
    return stringToRole[text];
}

func GetRoleString(role UserRole) string {
    return roleToString[role];
}

func GetAllRoles() map[UserRole]string {
    return roleToString;
}

func GetAllRoleStrings() map[string]UserRole {
    return stringToRole;
}

func (this UserRole) MarshalJSON() ([]byte, error) {
    buffer := bytes.NewBufferString(`"`);
    buffer.WriteString(roleToString[this]);
    buffer.WriteString(`"`);
    return buffer.Bytes(), nil;
}

func (this *UserRole) UnmarshalJSON(data []byte) error {
    var temp string;

    err := json.Unmarshal(data, &temp);
    if (err != nil) {
        return err;
    }

    temp = strings.ToLower(temp);

    var ok bool;
    *this, ok = stringToRole[temp];
    if (!ok) {
        *this = RoleUnknown;
        return fmt.Errorf("RoleUnknown UserRole value: '%s'.", temp);
    }

    return nil;
}
