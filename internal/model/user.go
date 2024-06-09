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

// Get a server user representation of this data.
// Passwords will NOT be converted into tokens (as the source salt is unknown).
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

	if this.Course != "" {
		user.Roles[this.Course] = GetCourseUserRole(this.CourseRole)

		if this.CourseLMSID != "" {
			user.LMSIDs[this.Course] = this.CourseLMSID
		}
	}

	return user, user.Validate()
}

// Does this data have server-level user information?
func (this *RawUserData) HasServerInfo() bool {
	return (this.Name != "") || (this.Role != "")
}

// Does this data have course-level user information?
func (this *RawUserData) HasCourseInfo() bool {
	return (this.Course != "") || (this.CourseRole != "") || (this.CourseLMSID != "")
}
