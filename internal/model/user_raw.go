package model

import (
	"github.com/edulinq/autograder/internal/util"
)

// Raw/dirty data for a user.
// This struct can be directly embedded for Kong arguments.
type RawUserData struct {
	Email       string `json:"email" help:"Email for the user." arg:"" required:""`
	Name        string `json:"name" help:"Name for the user."`
	Role        string `json:"role" help:"Server role for the user. Defaults to 'user'." default:"user"`
	Pass        string `json:"pass" help:"Password for the user. Defaults to a random string (will be output)."`
	Course      string `json:"course" help:"Optional ID of course to enroll user in."`
	CourseRole  string `json:"course-role" help:"Role for the new user in the specified course. Defaults to 'student'." default:"student"`
	CourseLMSID string `json:"course-lms-id" help:"LMS ID for the new user in the specified course."`
}

// Raw/dirty data for a course user.
// This struct is used for raw data coming from a single course.
type RawCourseUserData struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	CourseRole  string `json:"course-role"`
	CourseLMSID string `json:"course-lms-id"`
}

func (this *RawCourseUserData) ToRawUserData(course *Course) *RawUserData {
	rawUserData := &RawUserData{
		Email:       this.Email,
		Name:        this.Name,
		Role:        "",
		Pass:        "",
		Course:      course.GetID(),
		CourseRole:  this.CourseRole,
		CourseLMSID: this.CourseLMSID,
	}

	return rawUserData
}

func ToRawUserDatas(rawCourseUsers []*RawCourseUserData, course *Course) []*RawUserData {
	rawUsers := make([]*RawUserData, 0, len(rawCourseUsers))

	for _, rawCourseUser := range rawCourseUsers {
		if rawCourseUser != nil {
			rawUsers = append(rawUsers, rawCourseUser.ToRawUserData(course))
		}
	}

	return rawUsers
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
