package model

import (
	"fmt"
	"strings"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var COURSE_USER_ROW_COLUMNS []string = []string{"email", "name", "role", "lms-id"}

// CourseUsers represent users enrolled in a course (in any role including owner).
// They only contain a users information that is relevant to the course.
// Pointer fields indicate optional fields.
type CourseUser struct {
	Email string         `json:"email"`
	Name  *string        `json:"name"`
	Role  CourseUserRole `json:"role"`
	LMSID *string        `json:"lms-id"`
}

func NewCourseUser(email string, name *string, role CourseUserRole, lmsID *string) (*CourseUser, error) {
	courseUser := &CourseUser{
		Email: email,
		Name:  name,
		Role:  role,
		LMSID: lmsID,
	}

	return courseUser, courseUser.Validate()
}

func (this *CourseUser) Validate() error {
	this.Email = strings.TrimSpace(this.Email)
	if this.Email == "" {
		return fmt.Errorf("User email is empty.")
	}

	if this.Name != nil {
		name := strings.TrimSpace(*this.Name)
		this.Name = &name
	}

	if this.Role == CourseRoleUnknown {
		return fmt.Errorf("User '%s' has an unknown role. All users must have a definite role.", this.Email)
	}

	if this.LMSID != nil {
		lmsID := strings.TrimSpace(*this.LMSID)
		this.LMSID = &lmsID
	}

	return nil
}

func (this *CourseUser) LogValue() []*log.Attr {
	return []*log.Attr{log.NewUserAttr(this.Email)}
}

func (this *CourseUser) GetName(fallback bool) string {
	name := ""

	if this.Name != nil {
		name = *this.Name
	}

	if fallback && (name == "") {
		name = this.Email
	}

	return name
}

func (this *CourseUser) GetDisplayName() string {
	return this.GetName(true)
}

// Get a string (not pointer) representation of this user's LMS ID.
func (this *CourseUser) GetLMSID() string {
	if this.LMSID == nil {
		return ""
	}

	return *this.LMSID
}

// Note that this function is potentially dangerous because
// we are converting from a type with less information into
// a type with more information.
func (this *CourseUser) ToServerUser(courseID string) (*ServerUser, error) {
	serverUser := &ServerUser{
		Email:      this.Email,
		Name:       this.Name,
		CourseInfo: map[string]*UserCourseInfo{courseID: &UserCourseInfo{Role: this.Role, LMSID: this.LMSID}},
	}

	return serverUser, serverUser.validate(false)
}

func (this *CourseUser) MustToRow() []string {
	return []string{
		this.Email,
		util.PointerToString(this.Name),
		this.Role.String(),
		util.PointerToString(this.LMSID),
	}
}
