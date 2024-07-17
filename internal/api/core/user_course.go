package core

import (
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

// An API-safe representation of a course user.
type CourseUserInfo struct {
	Email string               `json:"email"`
	Name  string               `json:"name"`
	Role  model.CourseUserRole `json:"role"`
	LMSID string               `json:"lms-id"`
}

func NewCourseUserInfo(user *model.CourseUser) *CourseUserInfo {
	return &CourseUserInfo{
		Email: user.Email,
		Name:  user.GetDisplayName(),
		Role:  user.Role,
		LMSID: user.GetLMSID(),
	}
}

func NewCourseUserInfos(users []*model.CourseUser) []*CourseUserInfo {
	result := make([]*CourseUserInfo, 0, len(users))
	for _, user := range users {
		result = append(result, NewCourseUserInfo(user))
	}

	slices.SortFunc(result, CompareCourseUserInfoPointer)

	return result
}

func CompareCourseUserInfoPointer(a *CourseUserInfo, b *CourseUserInfo) int {
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
func CourseUserInfoFromMap(data map[string]any) *CourseUserInfo {
	return &CourseUserInfo{
		Email: data["email"].(string),
		Name:  data["name"].(string),
		Role:  model.GetCourseUserRole(data["role"].(string)),
		LMSID: data["lms-id"].(string),
	}
}

func CompareCourseUserInfo(a CourseUserInfo, b CourseUserInfo) int {
	return strings.Compare(a.Email, b.Email)
}
