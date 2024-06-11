package core

// How to represent users in API responses.

import (
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

type UserInfo struct {
	Email string               `json:"email"`
	Name  string               `json:"name"`
	Role  model.CourseUserRole `json:"role"`
	LMSID string               `json:"lms-id"`
}

func NewUserInfo(user *model.CourseUser) *UserInfo {
	return &UserInfo{
		Email: user.Email,
		Name:  user.GetDisplayName(),
		Role:  user.Role,
		LMSID: user.GetLMSID(),
	}
}

func NewUserInfos(users []*model.CourseUser) []*UserInfo {
	result := make([]*UserInfo, 0, len(users))
	for _, user := range users {
		result = append(result, NewUserInfo(user))
	}

	slices.SortFunc(result, CompareUserInfoPointer)

	return result
}

func CompareUserInfoPointer(a *UserInfo, b *UserInfo) int {
	if a == b {
		return 0
	}

	if a == nil {
		return 1
	}

	if b == nil {
		return -1
	}

	return strings.Compare(a.Email, b.Email)
}

// Get user info from a generic map (like what an API response would have).
func UserInfoFromMap(data map[string]any) *UserInfo {
	return &UserInfo{
		Email: data["email"].(string),
		Name:  data["name"].(string),
		Role:  model.GetCourseUserRole(data["role"].(string)),
		LMSID: data["lms-id"].(string),
	}
}

func CompareUserInfo(a UserInfo, b UserInfo) int {
	return strings.Compare(a.Email, b.Email)
}
