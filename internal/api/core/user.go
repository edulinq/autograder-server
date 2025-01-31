package core

import (
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/model"
)

type UserInfoType string

const (
	ServerUserInfoType = "server"
	CourseUserInfoType = "course"
)

// This type must be embedded into any API-safe representation of a user.
type BaseUserInfo struct {
	Type  UserInfoType `json:"type"`
	Email string       `json:"email"`
	Name  string       `json:"name"`
}

// An API-safe representation of a server user.
// Embed the BaseUserInfo and use ServerUserInfoType as the type.
type ServerUserInfo struct {
	BaseUserInfo
	Role    model.ServerUserRole      `json:"role"`
	Courses map[string]EnrollmentInfo `json:"courses"`
}

// An API-safe representation of enrollment information.
type EnrollmentInfo struct {
	CourseID string               `json:"id"`
	Role     model.CourseUserRole `json:"role"`
}

// An API-safe representation of a course user.
// Embed the BaseUserInfo and use CourseUserInfoType as the type.
type CourseUserInfo struct {
	BaseUserInfo
	Role  model.CourseUserRole `json:"role"`
	LMSID string               `json:"lms-id"`
}

func NewServerUserInfo(user *model.ServerUser) *ServerUserInfo {
	info := &ServerUserInfo{
		BaseUserInfo: BaseUserInfo{
			Type:  ServerUserInfoType,
			Email: user.Email,
			Name:  user.GetDisplayName(),
		},
		Role:    user.Role,
		Courses: make(map[string]EnrollmentInfo, len(user.CourseInfo)),
	}

	for courseID, courseInfo := range user.CourseInfo {
		info.Courses[courseID] = EnrollmentInfo{
			CourseID: courseID,
			Role:     courseInfo.Role,
		}
	}

	return info
}

func NewServerUserInfos(users []*model.ServerUser) []*ServerUserInfo {
	infos := make([]*ServerUserInfo, 0, len(users))
	for _, user := range users {
		infos = append(infos, NewServerUserInfo(user))
	}

	slices.SortFunc(infos, CompareServerUserInfoPointer)

	return infos
}

func CompareServerUserInfoPointer(a *ServerUserInfo, b *ServerUserInfo) int {
	if a == b {
		return 0
	}

	if a == nil {
		return 1
	}

	if b == nil {
		return -1
	}

	return CompareServerUserInfo(*a, *b)
}

func CompareServerUserInfo(a ServerUserInfo, b ServerUserInfo) int {
	return strings.Compare(a.Email, b.Email)
}

func NewCourseUserInfo(user *model.CourseUser) *CourseUserInfo {
	info := &CourseUserInfo{
		BaseUserInfo: BaseUserInfo{
			Type:  CourseUserInfoType,
			Email: user.Email,
			Name:  user.GetDisplayName(),
		},
		Role:  user.Role,
		LMSID: user.GetLMSID(),
	}

	return info
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

	return CompareCourseUserInfo(*a, *b)
}

func CompareCourseUserInfo(a CourseUserInfo, b CourseUserInfo) int {
	return strings.Compare(a.Email, b.Email)
}
