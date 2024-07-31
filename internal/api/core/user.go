package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type UserInfoType int

const (
	BaseUserInfoType UserInfoType = iota
	ServerUserInfoType
	CourseUserInfoType
)

// UserInfoInterface defines the methods that all user types must implement.
type UserInfoInterface interface {
	GetType() UserInfoType
}

// An API-safe representation of any user.
// UserInfo is the superclass that holds reference to any UserInfoInterface and basic user info.
type UserInfo struct {
	Email string            `json:"email"`
	Name  string            `json:"name"`
	Info  UserInfoInterface `json:"-"`
}

func NewUserInfo(email string, name string, info *UserInfoInterface) *UserInfo {
	return &UserInfo{
		Email: email,
		Name:  name,
		Info:  *info,
	}
}

func (userInfo *UserInfo) PrintType() {
	fmt.Printf("\nUserInfo type: '%d'.\n", userInfo.Info.GetType())
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

func CompareUserInfo(a UserInfo, b UserInfo) int {
	return strings.Compare(a.Email, b.Email)
}

// An API-safe representation of a server user.
type ServerUserInfo struct {
	UserInfo
	Role    model.ServerUserRole      `json:"role"`
	Courses map[string]EnrollmentInfo `json:"courses"`
}

type EnrollmentInfo struct {
	CourseID   string               `json:"id"`
	CourseName string               `json:"name"`
	Role       model.CourseUserRole `json:"role"`
}

func (serverUserInfo ServerUserInfo) GetType() UserInfoType {
	return ServerUserInfoType
}

func NewServerUserInfo(user *model.ServerUser) (*ServerUserInfo, error) {
	info := &ServerUserInfo{
		UserInfo: UserInfo{
			Email: user.Email,
			Name:  user.GetDisplayName(),
			Info:  nil,
		},
		Role:    user.Role,
		Courses: make(map[string]EnrollmentInfo, len(user.CourseInfo)),
	}

	// Set the Info field to the ServerUserInfo instance itself.
	info.Info = info

	for courseID, courseInfo := range user.CourseInfo {
		course, err := db.GetCourse(courseID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get course (%s) for user (%s).", courseID, user.Email)
		}

		info.Courses[course.GetID()] = EnrollmentInfo{
			CourseID:   course.GetID(),
			CourseName: course.GetName(),
			Role:       courseInfo.Role,
		}
	}

	return info, nil
}

func MustNewServerUserInfo(user *model.ServerUser) *ServerUserInfo {
	info, err := NewServerUserInfo(user)
	if err != nil {
		log.Fatal("Failed to convert server user to API info.", err, user)
	}

	return info
}

func NewServerUserInfos(users []*model.ServerUser) ([]*ServerUserInfo, error) {
	infos := make([]*ServerUserInfo, 0, len(users))
	for _, user := range users {
		info, err := NewServerUserInfo(user)
		if err != nil {
			return nil, fmt.Errorf("Failed to get server user info for user (%s).", user.Email)
		}

		infos = append(infos, info)
	}

	slices.SortFunc(infos, CompareServerUserInfoPointer)

	return infos, nil
}

func MustNewServerUserInfos(users []*model.ServerUser) []*ServerUserInfo {
	infos, err := NewServerUserInfos(users)
	if err != nil {
		log.Fatal("Failed to convert server users to API infos.", err, users)
	}

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

	return CompareUserInfoPointer(&a.UserInfo, &b.UserInfo)
}

func CompareServerUserInfo(a ServerUserInfo, b ServerUserInfo) int {
	return CompareUserInfo(a.UserInfo, b.UserInfo)
}

// An API-safe representation of a course user.
type CourseUserInfo struct {
	UserInfo
	Role  model.CourseUserRole `json:"role"`
	LMSID string               `json:"lms-id"`
}

func NewCourseUserInfo(user *model.CourseUser) *CourseUserInfo {
	info := &CourseUserInfo{
		UserInfo: UserInfo{
			Email: user.Email,
			Name:  user.GetDisplayName(),
			Info:  nil,
		},
		Role:  user.Role,
		LMSID: user.GetLMSID(),
	}

	// Set the Info field to the CourseUserInfo instance itself.
	info.Info = info

	return info
}

func (courseUserInfo CourseUserInfo) GetType() UserInfoType {
	return CourseUserInfoType
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

	return CompareUserInfoPointer(&a.UserInfo, &b.UserInfo)
}

func CompareCourseUserInfo(a CourseUserInfo, b CourseUserInfo) int {
	return CompareUserInfo(a.UserInfo, b.UserInfo)
}
