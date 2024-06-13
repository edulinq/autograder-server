package model

import (
	"github.com/edulinq/autograder/internal/util"
)

// Raw/dirty data for a user.
// This struct can be directly embedded for Kong arguments.
type RawUserData struct {
	Email       string `help:"Email for the user." arg:"" required:""`
	Name        string `help:"Name for the user."`
	Role        string `help:"Server role for the user. Defaults to 'user'." default:"user"`
	Pass        string `help:"Password for the user. Defaults to a random string (will be output)."`
	Course      string `help:"Optional ID of course to enroll user in."`
	CourseRole  string `help:"Role for the new user in the specified course. Defaults to 'student'." default:"student"`
	CourseLMSID string `help:"LMS ID for the new user in the specified course."`
}

// Get a server user representation of this data.
// Passwords will NOT be converted into tokens (as the source salt is unknown).
func (this *RawUserData) ToServerUser() (*ServerUser, error) {
	user := &ServerUser{
		Email: this.Email,
		Name:  nil,
		Role:  GetServerUserRole(this.Role),
		Salt:  nil,

		Tokens:     make([]*Token, 0),
		CourseInfo: make(map[string]*UserCourseInfo, 0),
	}

	if this.Name != "" {
		user.Name = util.StringPointer(this.Name)
	}

	if this.Course != "" {
		user.CourseInfo[this.Course] = &UserCourseInfo{
			Role: GetCourseUserRole(this.CourseRole),
		}

		if this.CourseLMSID != "" {
			user.CourseInfo[this.Course].LMSID = &this.CourseLMSID
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
