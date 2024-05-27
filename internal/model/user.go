package model

// It is expected that any password passed into functions here
// are already a hex encoding of a sha256 hash of the original cleartext
// see util.Sha256Hex().

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/email"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

// TEST -- Full tests on validate and upsert.

// TEST - Core users with shared functionality?

// TEST
const (
	DEFAULT_PASSWORD_LEN = 32
	SALT_LENGTH_BYTES    = 32

	ARGON2_KEY_LEN_BYTES = 32
	ARGON2_MEM_KB        = 64 * 1024
	ARGON2_THREADS       = 4
	ARGON2_TIME          = 1

	EMAIL_SLEEP_TIME = int64(1.5 * float64(time.Second))
)

// TEST - Split server and course into files?

// TEST
// Pointer fields indicate optional fields.
// Note that optional fields are not optional in all context.

// ServerUsers represent general users that exist on a server.
// They may or may not be enrolled in courses.
// ServerUsers should generally only be used for server-level activities,
// most grading-related code should be using CourseUsers.
type ServerUser struct {
	Email string  `json:"email"`
	Name  *string `json:"name"`
	Salt  *string `json:"salt"`

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

// CourseUsers represent users enrolled in a course (as students, graders, etc).
// They only contain a users information that is relevant to the course.
type CourseUser struct {
	Email string   `json:"email"`
	Name  *string  `json:"name"`
	Role  UserRole `json:"role"`
	LMSID *string  `json:"lms-id"`
}

func NewCourseUser(email string, name *string, role UserRole, lmsID *string) (*CourseUser, error) {
	courseUser := &CourseUser{
		Email: email,
		Name:  name,
		Role:  role,
		LMSID: lmsID,
	}

	return courseUser, courseUser.Validate()
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
		salt := strings.TrimSpace(*this.Salt)
		this.Salt = &salt
	}

	if this.Tokens == nil {
		this.Tokens = make([]string, 0)
	}

	for i, token := range this.Tokens {
		token = strings.ToLower(strings.TrimSpace(token))
		if token == "" {
			return fmt.Errorf("User '%s' has a token (index %d) that is empty.", this.Email, i)
		}

		if token == "" {
			return fmt.Errorf("User '%s' has a token (index %d) that is empty.", this.Email, i)
		}

		_, err := hex.DecodeString(token)
		if err != nil {
			return fmt.Errorf("User '%s' has a token (index %d) that is not proper hex: '%w'.", this.Email, i, err)
		}

		this.Tokens[i] = token
	}

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

func (this *CourseUser) Validate() error {
	this.Email = strings.TrimSpace(this.Email)
	if this.Email == "" {
		return fmt.Errorf("User email is empty.")
	}

	if this.Name != nil {
		name := strings.TrimSpace(*this.Name)
		this.Name = &name
	}

	if this.Role == RoleUnknown {
		return fmt.Errorf("User '%s' has an unknown role. All users must have a definite role.", this.Email)
	}

	if this.LMSID != nil {
		lmsID := strings.TrimSpace(*this.LMSID)
		this.LMSID = &lmsID
	}

	return nil
}

func (this *ServerUser) LogValue() []*log.Attr {
	return []*log.Attr{log.NewUserAttr(this.Email)}
}

func (this *CourseUser) LogValue() []*log.Attr {
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

func (this *CourseUser) GetServerUser(course *Course) (*ServerUser, error) {
	serverUser := &ServerUser{
		Email: this.Email,
		Name:  this.Name,
		Roles: map[string]UserRole{course.ID: this.Role},
	}

	if this.LMSID != nil {
		serverUser.LMSIDs = map[string]string{course.ID: *this.LMSID}
	}

	return serverUser, serverUser.Validate()
}

// Add information from the given user to this user.
// Everything but email can be added (the email of the two users must already match).
// Any given tokens will be added.
// Any Roles or LMSIDs will be upserted.
func (this *ServerUser) Upsert(other *ServerUser) {
	if other == nil {
		return
	}

	if this.Email != other.Email {
		return
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

		// Sort and removes duplicates.
		slices.Sort(this.Tokens)
		this.Tokens = slices.Compact(this.Tokens)
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
}

// TEST

// TEST
type User struct {
	Email string   `json:"email"`
	Name  string   `json:"name"`
	Role  UserRole `json:"role"`
	Pass  string   `json:"pass"`
	Salt  string   `json:"salt"`

	LMSID string `json:"lms-id"`
}

func NewUser(email string, name string, role UserRole) *User {
	return &User{
		Email: email,
		Name:  name,
		Role:  role,
	}
}

func (this *User) LogValue() []*log.Attr {
	return []*log.Attr{log.NewUserAttr(this.Email)}
}

// Sets the password and generates a new salt.
// The passed in passowrd should actually be a hash of the cleartext password.
func (this *User) SetPassword(hashPass string) error {
	salt, err := util.RandBytes(SALT_LENGTH_BYTES)
	if err != nil {
		return fmt.Errorf("Could not generate salt: '%w'.", err)
	}

	pass := generateHash(hashPass, salt)

	this.Salt = hex.EncodeToString(salt)
	this.Pass = hex.EncodeToString(pass)

	return nil
}

// Set a random passowrd, and return the cleartext (not hash) password.
func (this *User) SetRandomPassword() (string, error) {
	pass, err := util.RandHex(DEFAULT_PASSWORD_LEN)
	if err != nil {
		return "", fmt.Errorf("Failed to generate random password: '%s'.", err)
	}

	hashPass := util.Sha256HexFromString(pass)
	err = this.SetPassword(hashPass)
	if err != nil {
		return "", err
	}

	return pass, nil
}

// Return true if the password matches the hash, false otherwise.
// Any errors (which can only come from bad hex strings) will be logged and ignored (false will be returned).
func (this *User) CheckPassword(hashPass string) bool {
	thisHash, err := hex.DecodeString(this.Pass)
	if err != nil {
		log.Warn("Bad password hash for user.", err, this)
		return false
	}

	salt, err := hex.DecodeString(this.Salt)
	if err != nil {
		log.Warn("Bad salt for user.", err, this)
		return false
	}

	otherHash := generateHash(hashPass, salt)

	return (subtle.ConstantTimeCompare(thisHash, otherHash) == 1)
}

// Merge another user's information into this user (email will not be merged).
// Empty values will not be merged.
// Returns true if any changes were made.
func (this *User) Merge(other *User) bool {
	changed := false

	if (other.Name != "") && (this.Name != other.Name) {
		this.Name = other.Name
		changed = true
	}

	if (other.Pass != "") && (this.Pass != other.Pass) {
		this.Pass = other.Pass
		this.Salt = other.Salt
		changed = true
	}

	if (other.Role != RoleUnknown) && (this.Role != other.Role) {
		this.Role = other.Role
		changed = true
	}

	if (other.LMSID != "") && (this.LMSID != other.LMSID) {
		this.LMSID = other.LMSID
		changed = true
	}

	return changed
}

func generateHash(hashPass string, salt []byte) []byte {
	return argon2.IDKey([]byte(hashPass), salt, ARGON2_TIME, ARGON2_MEM_KB, ARGON2_THREADS, ARGON2_KEY_LEN_BYTES)
}

func SendUserAddEmail(course *Course, user *User, pass string, generatedPass bool, userExists bool, dryRun bool, sleep bool) error {
	subject, body := composeUserAddEmail(course, user.Email, pass, generatedPass, userExists)

	if dryRun {
		log.Info("Doing a dry run, user will not be emailed.", course, log.NewUserAttr(user.Email))
		log.Debug("Email not sent because of dry run.", course,
			log.NewAttr("address", user.Email), log.NewAttr("subject", subject), log.NewAttr("body", body))
		return nil
	}

	err := email.Send([]string{user.Email}, subject, body, false)
	if err != nil {
		log.Error("Failed to send email.", err, course, log.NewUserAttr(user.Email))
		return err
	}

	log.Info("Registration email sent.", course, log.NewUserAttr(user.Email))

	// Skip sleeping in testing mode.
	if sleep && !config.TESTING_MODE.Get() {
		time.Sleep(time.Duration(EMAIL_SLEEP_TIME))
	}

	return nil
}

func composeUserAddEmail(course *Course, address string, pass string, generatedPass bool, userExists bool) (string, string) {
	var subject string
	var body string

	if userExists {
		subject = fmt.Sprintf("Autograder %s -- User Password Changed", course.GetID())
		body =
			"Hello,\n" +
				fmt.Sprintf("\nThe password for '%s' has been changed for the course '%s'.\n", address, course.GetDisplayName())
	} else {
		subject = fmt.Sprintf("Autograder %s -- User Account Created", course.GetID())
		body =
			"Hello,\n" +
				fmt.Sprintf("\nAn autograder account with the username/email '%s' has been created for the course '%s'.\n", address, course.GetDisplayName()) +
				"Usage instructions will provided in class.\n"
	}

	if generatedPass {
		body += fmt.Sprintf("Your new password is '%s' (no quotes).\n", pass)
	}

	return subject, body
}

func ToRowHeaader(delim string) string {
	parts := []string{"email", "name", "role", "lms-id"}
	return strings.Join(parts, delim)
}

func (this *User) ToRow(delim string) string {
	parts := []string{this.Email, this.Name, this.Role.String(), this.LMSID}
	return strings.Join(parts, delim)
}

func UsersPointerEqual(a []*User, b []*User) bool {
	if len(a) != len(b) {
		return false
	}

	slices.SortFunc(a, UserPointerCompare)
	slices.SortFunc(b, UserPointerCompare)

	return slices.EqualFunc(a, b, UserPointerEqual)
}

func UserPointerEqual(a *User, b *User) bool {
	if a == b {
		return true
	}

	if (a == nil) || (b == nil) {
		return false
	}

	return (a.Email == b.Email) &&
		(a.Name == b.Name) &&
		(a.Role == b.Role) &&
		(a.LMSID == b.LMSID)
}

func UserPointerCompare(a *User, b *User) int {
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
