package model

import (
	"github.com/edulinq/autograder/internal/util"
)

// Server user roles represent a user's role within an autograder server instance.
type ServerUserRole int

// ServerRoleUnknown is the zero value and no user should have this role (it is a validation error).
// ServerRoleUser is for standard users (these users can even be owners of courses).
// ServerRoleCourseCreator is for users that can create courses, and administrate their OWN courses.
// ServerRoleAdmin is for users that can administrate ALL courses.
// ServerRoleOwner is for the top-level authorities (that are real users) on the server.
// ServerRoleRoot is not for an actual user (will be a validation error), but the authority given when running CMDs directly.
const (
	ServerRoleUnknown       ServerUserRole = 0
	ServerRoleUser                         = 10
	ServerRoleCourseCreator                = 20
	ServerRoleAdmin                        = 30
	ServerRoleOwner                        = 40
	ServerRoleRoot                         = 50
)

var serverRoleToString = map[ServerUserRole]string{
	ServerRoleUnknown:       "unknown",
	ServerRoleUser:          "user",
	ServerRoleCourseCreator: "creator",
	ServerRoleAdmin:         "admin",
	ServerRoleOwner:         "owner",
	ServerRoleRoot:          "root",
}

var stringToServerUserRole = map[string]ServerUserRole{
	"unknown": ServerRoleUnknown,
	"user":    ServerRoleUser,
	"creator": ServerRoleCourseCreator,
	"admin":   ServerRoleAdmin,
	"owner":   ServerRoleOwner,
	"root":    ServerRoleRoot,
}

func GetServerUserRole(text string) ServerUserRole {
	return stringToServerUserRole[text]
}

func GetServerUserRoleString(role ServerUserRole) string {
	return serverRoleToString[role]
}

func GetAllServerUserRoles() map[ServerUserRole]string {
	return serverRoleToString
}

func GetAllServerUserRoleSStrings() map[string]ServerUserRole {
	return stringToServerUserRole
}

func GetCommonServerUserRoleStrings() map[string]any {
	commonServerRoles := make(map[string]any, 0)
	for roleString, _ := range stringToServerUserRole {
		if roleString == "unknown" {
			continue
		}

		if roleString == "root" {
			continue
		}

		commonServerRoles[roleString] = nil
	}

	return commonServerRoles
}

func (this ServerUserRole) String() string {
	return serverRoleToString[this]
}

func (this ServerUserRole) MarshalJSON() ([]byte, error) {
	return util.MarshalEnum(this, serverRoleToString)
}

func (this *ServerUserRole) UnmarshalJSON(data []byte) error {
	value, err := util.UnmarshalEnum(data, stringToServerUserRole, true)
	if err == nil {
		*this = *value
	}

	return err
}
