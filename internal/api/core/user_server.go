package core

import (
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// An API-safe representation of a server user.
type ServerUserInfo struct {
	Email   string                    `json:"email"`
	Name    string                    `json:"name"`
	Role    model.ServerUserRole      `json:"role"`
	Courses map[string]EnrollmentInfo `json:"courses"`
}

type EnrollmentInfo struct {
	CourseID   string               `json:"id"`
	CourseName string               `json:"name"`
	Role       model.CourseUserRole `json:"role"`
}

func NewServerUserInfo(user *model.ServerUser) (*ServerUserInfo, error) {
	info := &ServerUserInfo{
		Email:   user.Email,
		Name:    user.GetDisplayName(),
		Role:    user.Role,
		Courses: make(map[string]EnrollmentInfo, len(user.CourseInfo)),
	}

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
	result := make([]*ServerUserInfo, 0, len(users))
	for _, user := range users {
		info, err := NewServerUserInfo(user)
		if err != nil {
			return nil, fmt.Errorf("Failed to get server user info for user (%s).", user.Email)
		}

		result = append(result, info)
	}

	slices.SortFunc(result, CompareServerUserInfoPointer)

	return result, nil
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

	return strings.Compare(a.Email, b.Email)
}

// Get server user info from a generic map (like what an API response would have).
func ServerUserInfoFromMap(data map[string]any) *ServerUserInfo {
	courses := data["courses"].(map[string]any)

	return &ServerUserInfo{
		Email:   data["email"].(string),
		Name:    data["name"].(string),
		Role:    model.GetServerUserRole(data["role"].(string)),
		Courses: *EnrollmentInfoFromMap(courses),
	}
}

// Get enrollment info from a generic map (like what an API response would have).
func EnrollmentInfoFromMap(data map[string]any) *map[string]EnrollmentInfo {
	enrollmentInfos := make(map[string]EnrollmentInfo)

	for name, course := range data {
		courseMap := course.(map[string]any)
		enrollmentInfos[name] = EnrollmentInfo{
			CourseID:   courseMap["id"].(string),
			CourseName: courseMap["name"].(string),
			Role:       model.GetCourseUserRole(courseMap["role"].(string)),
		}
	}

	return &enrollmentInfos
}

func CompareServerUserInfo(a ServerUserInfo, b ServerUserInfo) int {
	return strings.Compare(a.Email, b.Email)
}
