package model

import (
    "bytes"
    "encoding/json"
    "fmt"
    "strings"
)

type UserRole int;

const (
    Unknown UserRole = 0
    Other            = 10
    Student          = 20
    Grader           = 30
    Admin            = 40
    Owner            = 50
)

func (this UserRole) String() string {
    return roleToString[this]
}

var roleToString = map[UserRole]string{
    Owner:   "owner",
    Admin:   "admin",
    Grader:  "grader",
    Student: "student",
    Other:   "other",
    Unknown: "unknown",
}

var stringToRole = map[string]UserRole{
    "owner":   Owner,
    "admin":   Admin,
    "grader":  Grader,
    "student": Student,
    "other":   Other,
    "unknown": Unknown,
}

func GetRole(text string) UserRole {
    return stringToRole[text];
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
        *this = Unknown;
        return fmt.Errorf("Unknown UserRole value: '%s'.", temp);
    }

    return nil;
}
