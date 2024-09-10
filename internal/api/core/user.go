package core

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

type UserInfoType string

const (
	ServerUserInfoType = "Server"
	CourseUserInfoType = "Course"
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
	CourseID   string               `json:"id"`
	CourseName string               `json:"name"`
	Role       model.CourseUserRole `json:"role"`
}

// An API-safe representation of a course user.
// Embed the BaseUserInfo and use CourseUserInfoType as the type.
type CourseUserInfo struct {
	BaseUserInfo
	Role  model.CourseUserRole `json:"role"`
	LMSID string               `json:"lms-id"`
}

func NewServerUserInfo(user *model.ServerUser) (*ServerUserInfo, error) {
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
		course, err := db.GetCourse(courseID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get course '%s' for user '%s'.", courseID, user.Email)
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
	var errs error

	infos := make([]*ServerUserInfo, 0, len(users))
	for _, user := range users {
		info, err := NewServerUserInfo(user)
		if err != nil {
			err = fmt.Errorf("Failed to get server user info for user '%s': '%w'.", user.Email, err)
			errs = errors.Join(errs, err)
		}

		infos = append(infos, info)
	}

	if errs != nil {
		return nil, fmt.Errorf("Found errors while creating new server user infos: '%w'.", errs)
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
