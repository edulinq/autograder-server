package core

import (
	"fmt"

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
	CourseID   string               `json:"course-id"`
	CourseName string               `json:"course-name"`
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
