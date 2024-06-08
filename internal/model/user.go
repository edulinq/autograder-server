package model

import (
	"github.com/edulinq/autograder/internal/util"
)

// Raw/dirty data for a user.
type RawUserData struct {
	Email       string
	Name        string
	Role        string
	Pass        string
	Course      string
	CourseRole  string
	CourseLMSID string
}

func (this *RawUserData) ToServerUser() (*ServerUser, error) {
	user := &ServerUser{
		Email: this.Email,
		Name:  nil,
		Role:  GetServerUserRole(this.Role),
		Salt:  nil,

		Tokens: make([]*Token, 0),
		Roles:  make(map[string]CourseUserRole, 0),
		LMSIDs: make(map[string]string, 0),
	}

	if this.Name != "" {
		user.Name = util.StringPointer(this.Name)
	}

	if user.Role == ServerRoleUnknown {
		user.Role = ServerRoleUser
	}

	if this.Course != "" {
		user.Roles[this.Course] = GetCourseUserRole(this.CourseRole)

		if this.CourseLMSID != "" {
			user.LMSIDs[this.Course] = this.CourseLMSID
		}
	}

	return user, user.Validate()
}
