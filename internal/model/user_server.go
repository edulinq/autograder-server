package model

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var SERVER_USER_ROW_COLUMNS []string = []string{"email", "name", "server-role", "salt", "password", "tokens", "course-info"}

// ServerUsers represent general users that exist on a server.
// They may or may not be enrolled in courses.
// ServerUsers should generally only be used for server-level activities,
// most grading-related code should be using CourseUsers.
// Pointer fields indicate optional fields.
// Note that optional fields are not optional in all context,
// e.g., a salt may not be required when updating a user but it is when authenticating.
type ServerUser struct {
	Email string         `json:"email"`
	Name  *string        `json:"name"`
	Role  ServerUserRole `json:"server-role"`

	// Salts shuold be hex strings.
	Salt *string `json:"salt"`

	// May be nil if the user does not have a password.
	Password *Token `json:"password"`

	// May be nil/empty if not retrieved from the database,
	// will always be non-nil after validation, but may be empty.
	// There should be no duplicates.
	Tokens []*Token `json:"tokens"`

	// May be nil/empty if not retrieved from the database,
	// will always be non-nil after validation.
	// Keyed by the course id.
	CourseInfo map[string]*UserCourseInfo `json:"course-info"`
}

var RootUserEmail = "root"

type UserCourseInfo struct {
	Role  CourseUserRole `json:"role"`
	LMSID *string        `json:"lms-id"`
}

func (this *ServerUser) Validate() error {
	this.Email = strings.TrimSpace(this.Email)
	if this.Email == "" {
		return fmt.Errorf("User email is empty.")
	}

	if this.Email != RootUserEmail && !strings.Contains(this.Email, "@") {
		return fmt.Errorf("User email '%s' has an invalid format.", this.Email)
	}

	if this.Name != nil {
		name := strings.TrimSpace(*this.Name)
		this.Name = &name

		if name == "" {
			return fmt.Errorf("User '%s' has an empty (but not nil) name.", this.Email)
		}
	}

	if this.Role == ServerRoleRoot && this.Email != RootUserEmail {
		return fmt.Errorf("User '%s' has a root server role. Normal users are not allowed to have this role.", this.Email)
	}

	if this.Salt != nil {
		salt := strings.ToLower(strings.TrimSpace(*this.Salt))
		this.Salt = &salt

		_, err := hex.DecodeString(*this.Salt)
		if err != nil {
			return fmt.Errorf("User '%s' has a salt that is not proper hex: '%w'.", this.Email, err)
		}
	}

	if this.Password != nil {
		if this.Password.Source != TokenSourcePassword {
			return fmt.Errorf("User '%s' has a password with a non-password source '%s'.", this.Email, this.Password.Source)
		}
	}

	if this.Tokens == nil {
		this.Tokens = make([]*Token, 0)
	}

	for i, token := range this.Tokens {
		err := token.Validate()
		if err != nil {
			return fmt.Errorf("User '%s' has a token (index %d) that is invalid: '%w'.", this.Email, i, err)
		}

		if token.Source == TokenSourcePassword {
			return fmt.Errorf("User '%s' has a token (index %d) that is marked as a password.", this.Email, i)
		}
	}

	this.compactTokens()

	if this.CourseInfo == nil {
		this.CourseInfo = make(map[string]*UserCourseInfo, 0)
	}

	newCourseInfo := make(map[string]*UserCourseInfo, len(this.CourseInfo))
	for courseID, info := range this.CourseInfo {
		newCourseID, err := common.ValidateID(strings.TrimSpace(courseID))
		if err != nil {
			return fmt.Errorf("User '%s' has a course info with invalid course id '%s': '%w'.", this.Email, courseID, err)
		}

		if info == nil {
			return fmt.Errorf("User '%s' has a nil course info '%s'.", this.Email, courseID)
		}

		err = info.Validate()
		if err != nil {
			return fmt.Errorf("User '%s' has an invalid course info '%s': '%w'.", this.Email, courseID, err)
		}

		newCourseInfo[newCourseID] = info
	}

	this.CourseInfo = newCourseInfo

	return nil
}

func (this *ServerUser) LogValue() []*log.Attr {
	return []*log.Attr{log.NewUserAttr(this.Email)}
}

func (this *ServerUser) GetName(fallback bool) string {
	name := ""

	if this.Name != nil {
		name = *this.Name
	}

	if fallback && (name == "") {
		name = this.Email
	}

	return name
}

func (this *ServerUser) GetDisplayName() string {
	return this.GetName(true)
}

func (this *ServerUser) GetCourseRole(courseID string) CourseUserRole {
	info, exists := this.CourseInfo[courseID]
	if !exists {
		return CourseRoleUnknown
	}

	return info.Role
}

// Convert this server user into a course user for the specific course.
// Set escalateServerAdmin to true if high level server users should be escalated to a course owner.
// Will return (nil, nil) if the user is not escalated and not enrolled in the given course.
func (this *ServerUser) ToCourseUser(courseID string, escalateServerAdmin bool) (*CourseUser, error) {
	escalate := escalateServerAdmin && (this.Role >= ServerRoleAdmin)

	info, enrolled := this.CourseInfo[courseID]
	if !enrolled && !escalate {
		return nil, nil
	}

	courseUser := &CourseUser{
		Email: this.Email,
		Name:  this.Name,
	}

	if enrolled {
		courseUser.Role = info.Role
		courseUser.LMSID = info.LMSID
	}

	if escalate {
		courseUser.Role = CourseRoleOwner
	}

	return courseUser, courseUser.Validate()
}

// Add information from the given user to this user.
// Everything but email can be added (the email of the two users must already match).
// Any given tokens will be added.
// Any course information will be upserted.
// After all merging, this user will be validated.
// Nothing can be removed in a merge.
// The returned boolean indicates if the context user was changed at all.
func (this *ServerUser) Merge(other *ServerUser) (bool, error) {
	if other == nil {
		return false, fmt.Errorf("Cannot merge with nil user.")
	}

	if this.Email != other.Email {
		return false, fmt.Errorf("Cannot merge users with different emails ('%s' and '%s').", this.Email, other.Email)
	}

	changed := false
	numTokens := len(this.Tokens)

	if (other.Name != nil) && ((this.Name == nil) || (*this.Name != *other.Name)) {
		changed = true
		newName := *other.Name
		this.Name = &newName
	}

	if (other.Role != ServerRoleUnknown) && (this.Role != other.Role) {
		changed = true
		this.Role = other.Role
	}

	if (other.Salt != nil) && ((this.Salt == nil) || (*this.Salt != *other.Salt)) {
		changed = true
		newSalt := *other.Salt
		this.Salt = &newSalt
	}

	if (other.Password != nil) && ((this.Password == nil) || !TokenPointerEqual(this.Password, other.Password)) {
		changed = true
		this.Password = other.Password.Clone()
	}

	if other.Tokens != nil {
		for _, token := range other.Tokens {
			this.Tokens = append(this.Tokens, token.Clone())
		}
	}

	if other.CourseInfo != nil {
		for courseID, otherInfo := range other.CourseInfo {
			info, exists := this.CourseInfo[courseID]
			if !exists {
				info = &UserCourseInfo{}
				changed = true
			}

			changed = info.Merge(otherInfo) || changed
			this.CourseInfo[courseID] = info
		}
	}

	err := this.Validate()
	if err != nil {
		return false, err
	}

	// Tokens will be sorted and compacted during validation.
	// Confirm the number of tokens (since they are addative in merges).
	changed = changed || (numTokens != len(this.Tokens))

	return changed, nil
}

// Set the password for a user.
// Note that this will replace the current password token (if it exists).
// Return true if the new password was set.
// The only way for (false, nil) is on a duplicate password.
func (this *ServerUser) SetPassword(password string) (bool, error) {
	if this.Salt == nil {
		return false, fmt.Errorf("User '%s' does not have a salt, and therefore cannot have a password.", this.Email)
	}

	newToken, err := NewToken(password, *this.Salt, TokenSourcePassword, TOKEN_PASSWORD_NAME)
	if err != nil {
		return false, fmt.Errorf("Failed to create token for user '%s' password.", this.Email)
	}

	// Check for duplicate password.
	if TokenPointerEqual(this.Password, newToken) {
		return false, nil
	}

	this.Password = newToken

	return true, nil
}

// Sort tokens and remove any duplicates.
func (this *ServerUser) compactTokens() {
	slices.SortFunc(this.Tokens, TokenPointerCompare)
	this.Tokens = slices.CompactFunc(this.Tokens, TokenPointerEqual)
}

func (this *ServerUser) SetRandomPassword() (string, error) {
	cleartext, err := RandomCleartext()
	if err != nil {
		return "", fmt.Errorf("User '%s' failed to generate random text for password: '%w'.", this.Email, err)
	}

	_, err = this.SetPassword(util.Sha256HexFromString(cleartext))
	if err != nil {
		return "", fmt.Errorf("User '%s' failed to set random password: '%w'.", this.Email, err)
	}

	return cleartext, nil
}

func (this *ServerUser) CreateRandomToken(name string, source TokenSource) (*Token, string, error) {
	if this.Salt == nil {
		return nil, "", fmt.Errorf("User '%s' does not have a salt, and therefore cannot have a token.", this.Email)
	}

	cleartext, token, err := NewRandomToken(*this.Salt, source, name)
	if err != nil {
		return nil, "", fmt.Errorf("User '%s' failed to generate a random token: '%w'.", this.Email, err)
	}

	this.Tokens = append(this.Tokens, token)
	this.compactTokens()

	return token, cleartext, nil
}

// Attempt to authenticate this user with the provided text.
// True will be returned if any of the tokens match.
func (this *ServerUser) Auth(input string) (bool, error) {
	var match bool = false
	var errs error = nil

	if this.Salt == nil {
		return false, fmt.Errorf("User '%s' has no salt. Cannot auth.", this.Email)
	}

	// Make sure that the password and all tokens are checked so we are not vulnerable to timing attacks.

	if this.Password != nil {
		tokenMatch, err := this.Password.Check(input, *this.Salt)
		errs = errors.Join(errs, err)
		match = match || tokenMatch
	}

	for _, token := range this.Tokens {
		tokenMatch, err := token.Check(input, *this.Salt)
		errs = errors.Join(errs, err)
		match = match || tokenMatch
	}

	if errs != nil {
		return false, errs
	}

	return match, nil
}

// Deep copy this user (which should already be validated).
func (this *ServerUser) Clone() *ServerUser {
	var password *Token = nil
	if this.Password != nil {
		password = this.Password.Clone()
	}

	tokens := make([]*Token, 0, len(this.Tokens))
	for _, token := range this.Tokens {
		tokens = append(tokens, token.Clone())
	}

	courseInfo := make(map[string]*UserCourseInfo, len(this.CourseInfo))
	for courseID, info := range this.CourseInfo {
		courseInfo[courseID] = info.Clone()
	}

	return &ServerUser{
		Email:      this.Email,
		Name:       this.Name,
		Role:       this.Role,
		Salt:       this.Salt,
		Password:   password,
		Tokens:     tokens,
		CourseInfo: courseInfo,
	}
}

func (this *ServerUser) MustToRow() []string {
	tokens := make([]string, 0, len(this.Tokens))
	for _, token := range this.Tokens {
		tokens = append(tokens, fmt.Sprintf("%s (%s)", token.Name, string(token.Source)))
	}

	password := "<nil>"
	if this.Password != nil {
		password = "<exists>"
	}

	return []string{
		this.Email,
		util.PointerToString(this.Name),
		this.Role.String(),
		util.PointerToString(this.Salt),
		password,
		util.MustToJSON(tokens),
		util.MustToJSON(this.CourseInfo),
	}
}

func (this *ServerUser) GetCourses() []string {
	courses := make([]string, 0, len(this.CourseInfo))
	for course, _ := range this.CourseInfo {
		courses = append(courses, course)
	}
	return courses
}

func (this *UserCourseInfo) GetLMSID() string {
	if this == nil {
		return ""
	}

	if this.LMSID == nil {
		return ""
	}

	return *this.LMSID
}

func (this *UserCourseInfo) Validate() error {
	if this.Role == CourseRoleUnknown {
		return fmt.Errorf("Unknown course role.")
	}

	if this.LMSID != nil {
		lmsID := strings.TrimSpace(*this.LMSID)
		if lmsID == "" {
			this.LMSID = nil
		} else {
			this.LMSID = &lmsID
		}
	}

	return nil
}

func (this *UserCourseInfo) Merge(other *UserCourseInfo) bool {
	if (this == nil) || (other == nil) {
		return false
	}

	changed := false

	if (other.Role != CourseRoleUnknown) && (this.Role != other.Role) {
		this.Role = other.Role
		changed = true
	}

	lmsID := util.PointerToString(this.LMSID)
	otherLMSID := util.PointerToString(other.LMSID)

	if (other.LMSID != nil) && (lmsID != otherLMSID) {
		this.LMSID = other.LMSID
		changed = true
	}

	return changed
}

func (this *UserCourseInfo) Clone() *UserCourseInfo {
	return &UserCourseInfo{
		Role:  this.Role,
		LMSID: this.LMSID,
	}
}
