package model

import (
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

var SERVER_USER_ROW_COLUMNS []string = []string{"email", "name", "server-role", "salt", "tokens", "course-roles", "lms-ids"}

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

	// May be nil/empty if not retrieved from the database,
	// will always be non-nil after validation, but may be empty.
	// There should be no duplicates and at most one token with the TokenSourcePassword source,
	// i.e., users can have many tokens, but only one derived from a password.
	Tokens []*Token `json:"tokens"`

	// May be nil/empty if not retrieved from the database,
	// will always be non-nil after validation.
	// Keyed by the course id.
	Roles  map[string]CourseUserRole `json:"course-roles"`
	LMSIDs map[string]string         `json:"lms-ids"`
}

func (this *ServerUser) Validate() error {
	this.Email = strings.TrimSpace(this.Email)
	if this.Email == "" {
		return fmt.Errorf("User email is empty.")
	}

	if this.Name != nil {
		name := strings.TrimSpace(*this.Name)
		this.Name = &name

		if name == "" {
			return fmt.Errorf("User '%s' has an empty (but not nil) name.", this.Email)
		}
	}

	// TEST
	if this.Role == ServerRoleRoot {
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

	if this.Tokens == nil {
		this.Tokens = make([]*Token, 0)
	}

	for i, token := range this.Tokens {
		err := token.Validate()
		if err != nil {
			return fmt.Errorf("User '%s' has a token (index %d) that is invalid: '%w'.", this.Email, i, err)
		}
	}

	slices.SortFunc(this.Tokens, TokenPointerCompare)
	this.Tokens = slices.CompactFunc(this.Tokens, TokenPointerEqual)

	if this.Roles == nil {
		this.Roles = make(map[string]CourseUserRole, 0)
	}

	newRoles := make(map[string]CourseUserRole, len(this.Roles))
	for courseID, role := range this.Roles {
		newCourseID, err := common.ValidateID(strings.TrimSpace(courseID))
		if err != nil {
			return fmt.Errorf("User '%s' has a role with invalid course id '%s': '%w'.", this.Email, courseID, err)
		}

		if role == RoleUnknown {
			return fmt.Errorf("User '%s' has an unknown role for course '%s'. All users must have a definite role.", this.Email, newCourseID)
		}

		newRoles[newCourseID] = role
	}

	this.Roles = newRoles

	if this.LMSIDs == nil {
		this.LMSIDs = make(map[string]string, 0)
	}

	newLMSIDs := make(map[string]string, len(this.LMSIDs))
	for courseID, lmsID := range this.LMSIDs {
		newCourseID, err := common.ValidateID(strings.TrimSpace(courseID))
		if err != nil {
			return fmt.Errorf("User '%s' has an LMS id with invalid course id '%s': '%w'.", this.Email, courseID, err)
		}

		lmsID = strings.TrimSpace(lmsID)
		if lmsID == "" {
			return fmt.Errorf("User '%s' has an empty LMS id for course '%s'.", this.Email, newCourseID)
		}

		newLMSIDs[newCourseID] = lmsID
	}

	this.LMSIDs = newLMSIDs

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

// Convert this server user into a course user for the specific course.
// Will return (nil, nil) if the user is not enrolled in the given course.
func (this *ServerUser) ToCourseUser(courseID string) (*CourseUser, error) {
	role, exists := this.Roles[courseID]
	if !exists {
		return nil, nil
	}

	var lmsID *string = nil
	lmsIDText, exists := this.LMSIDs[courseID]
	if exists {
		lmsID = &lmsIDText
	}

	courseUser := &CourseUser{
		Email: this.Email,
		Name:  this.Name,
		Role:  role,
		LMSID: lmsID,
	}

	return courseUser, courseUser.Validate()
}

// Add information from the given user to this user.
// Everything but email can be added (the email of the two users must already match).
// Any given tokens will be added.
// Any Roles or LMSIDs will be upserted.
// After all merging, this user will be validated.
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

	if other.Tokens != nil {
		for _, token := range other.Tokens {
			this.Tokens = append(this.Tokens, token)
		}
	}

	if other.Roles != nil {
		for key, newRole := range other.Roles {
			if this.Roles[key] != newRole {
				changed = true
				this.Roles[key] = newRole
			}
		}
	}

	if other.LMSIDs != nil {
		for key, lmsID := range other.LMSIDs {
			if this.LMSIDs[key] != lmsID {
				changed = true
				this.LMSIDs[key] = lmsID
			}
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

	oldIndex := -1
	for i, token := range this.Tokens {
		if token.Source == TokenSourcePassword {
			// Check for duplicates.
			if TokenPointerEqual(token, newToken) {
				return false, nil
			}

			oldIndex = i
			break
		}
	}

	if oldIndex == -1 {
		this.Tokens = append(this.Tokens, newToken)
	} else {
		this.Tokens[oldIndex] = newToken
	}

	slices.SortFunc(this.Tokens, TokenPointerCompare)
	this.Tokens = slices.CompactFunc(this.Tokens, TokenPointerEqual)

	return true, nil
}

func (this *ServerUser) SetRandomPassword() (string, error) {
	cleartext, err := RandomCleartext()
	if err != nil {
		return "", fmt.Errorf("User '%s' failed to generate random text for password: '%w'.", this.Email, err)
	}

	_, err = this.SetPassword(cleartext)
	if err != nil {
		return "", fmt.Errorf("User '%s' failed to set random passowrd: '%w'.", this.Email, err)
	}

	return cleartext, nil
}

// Deep copy this user (which should already be validated).
func (this *ServerUser) Clone() *ServerUser {
	tokens := make([]*Token, 0, len(this.Tokens))
	for _, token := range this.Tokens {
		tokens = append(tokens, token.Clone())
	}

	return &ServerUser{
		Email:  this.Email,
		Name:   this.Name,
		Role:   this.Role,
		Salt:   this.Salt,
		Tokens: tokens,
		Roles:  maps.Clone(this.Roles),
		LMSIDs: maps.Clone(this.LMSIDs),
	}
}

func (this *ServerUser) MustToRow() []string {
	tokens := make([]string, 0, len(this.Tokens))
	for _, token := range this.Tokens {
		tokens = append(tokens, fmt.Sprintf("%s (%s)", token.Name, string(token.Source)))
	}

	return []string{
		this.Email,
		util.PointerToString(this.Name),
		this.Role.String(),
		util.PointerToString(this.Salt),
		util.MustToJSON(tokens),
		util.MustToJSON(this.Roles),
		util.MustToJSON(this.LMSIDs),
	}
}

func (this *ServerUser) GetCourses() []string {
	courses := make([]string, 0, len(this.Roles))
	for course, _ := range this.Roles {
		courses = append(courses, course)
	}
	return courses
}
