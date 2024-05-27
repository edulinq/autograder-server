package model

import (
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
)

// TEST -- Full tests on validate and upsert.

// ServerUsers represent general users that exist on a server.
// They may or may not be enrolled in courses.
// ServerUsers should generally only be used for server-level activities,
// most grading-related code should be using CourseUsers.
// Pointer fields indicate optional fields.
// Note that optional fields are not optional in all context,
// e.g., a salt may not be required when updating a user but it is when authenticating.
type ServerUser struct {
	Email string  `json:"email"`
	Name  *string `json:"name"`
	// Salts shuold be hex strings.
	Salt *string `json:"salt"`

	// May be nil/empty if not retrieved from the database,
	// will always be non-nil after validation.
	// Tokens shuold be hex strings.
	Tokens []string `json:"tokens"`

	// May be nil/empty if not retrieved from the database,
	// will always be non-nil after validation.
	// Keyed by the course id.
	Roles  map[string]UserRole `json:"roles"`
	LMSIDs map[string]string   `json:"lms-ids"`
}

func (this *ServerUser) Validate() error {
	this.Email = strings.TrimSpace(this.Email)
	if this.Email == "" {
		return fmt.Errorf("User email is empty.")
	}

	if this.Name != nil {
		name := strings.TrimSpace(*this.Name)
		this.Name = &name
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
		this.Tokens = make([]string, 0)
	}

	for i, token := range this.Tokens {
		token = strings.ToLower(strings.TrimSpace(token))
		if token == "" {
			return fmt.Errorf("User '%s' has a token (index %d) that is empty.", this.Email, i)
		}

		_, err := hex.DecodeString(token)
		if err != nil {
			return fmt.Errorf("User '%s' has a token (index %d) that is not proper hex: '%w'.", this.Email, i, err)
		}

		this.Tokens[i] = token
	}

	// Sort and removes duplicates.
	slices.Sort(this.Tokens)
	this.Tokens = slices.Compact(this.Tokens)

	if this.Roles == nil {
		this.Roles = make(map[string]UserRole, 0)
	}

	newRoles := make(map[string]UserRole, len(this.Roles))
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
func (this *ServerUser) GetCourseUser(course *Course) (*CourseUser, error) {
	role, exists := this.Roles[course.ID]
	if !exists {
		return nil, nil
	}

	var lmsID *string = nil
	lmsIDText, exists := this.LMSIDs[course.ID]
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
func (this *ServerUser) Merge(other *ServerUser) error {
	if other == nil {
		return nil
	}

	if this.Email != other.Email {
		return nil
	}

	if other.Name != nil {
		this.Name = other.Name
	}

	if other.Salt != nil {
		this.Salt = other.Salt
	}

	if other.Tokens != nil {
		for _, token := range other.Tokens {
			this.Tokens = append(this.Tokens, token)
		}
	}

	if other.Roles != nil {
		if this.Roles == nil {
			this.Roles = make(map[string]UserRole, len(other.Roles))
		}

		for key, value := range other.Roles {
			this.Roles[key] = value
		}
	}

	if other.LMSIDs != nil {
		if this.LMSIDs == nil {
			this.LMSIDs = make(map[string]string, len(other.LMSIDs))
		}

		for key, value := range other.LMSIDs {
			this.LMSIDs[key] = value
		}
	}

	return this.Validate()
}
