package model

import (
    "bytes"
    "encoding/json"
    "fmt"
    "strings"
)

type UserRole int;

const (
    Owner UserRole = iota
    Admin
    Grader
    Student
    Other
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
}

var stringToRole = map[string]UserRole{
    "owner":   Owner,
    "admin":   Admin,
    "grader":  Grader,
    "student": Student,
    "other":   Other,
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
        *this = Other;
        return fmt.Errorf("Unknown UserRole value: '%s'.", temp);
    }

    return nil;
}
