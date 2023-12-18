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

    "github.com/rs/zerolog/log"
    "golang.org/x/crypto/argon2"

    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/email"
    "github.com/eriq-augustine/autograder/util"
)

const (
    DEFAULT_PASSWORD_LEN = 32;
    SALT_LENGTH_BYTES = 16;

    ARGON2_KEY_LEN_BYTES = 32;
    ARGON2_MEM_KB = 64 * 1024;
    ARGON2_THREADS = 4;
    ARGON2_TIME = 1;

    EMAIL_SLEEP_TIME = int64(1.5 * float64(time.Second));
)

type User struct {
    Email string `json:"email"`
    Name string `json:"name"`
    Role UserRole `json:"role"`
    Pass string `json:"pass"`
    Salt string `json:"salt"`

    LMSID string `json:"lms-id"`
}

func NewUser(email string, name string, role UserRole) *User {
    return &User{
        Email: email,
        Name: name,
        Role: role,
    };
}

// Sets the password and generates a new salt.
// The passed in passowrd should actually be a hash of the cleartext password.
func (this *User) SetPassword(hashPass string) error {
    salt, err := util.RandBytes(SALT_LENGTH_BYTES);
    if (err != nil) {
        return fmt.Errorf("Could not generate salt: '%w'.", err);
    }

    pass := generateHash(hashPass, salt);

    this.Salt = hex.EncodeToString(salt);
    this.Pass = hex.EncodeToString(pass);

    return nil;
}

// Set a random passowrd, and return the cleartext (not hash) password.
func (this *User) SetRandomPassword() (string, error) {
    pass, err := util.RandHex(DEFAULT_PASSWORD_LEN)
    if (err != nil) {
        return "", fmt.Errorf("Failed to generate random password: '%s'.", err);
    }

    hashPass := util.Sha256HexFromString(pass);
    err = this.SetPassword(hashPass);
    if (err != nil) {
        return "", err;
    }

    return pass, nil;
}

// Return true if the password matches the hash, false otherwise.
// Any errors (which can only come from bad hex strings) will be logged and ignored (false will be returned).
func (this *User) CheckPassword(hashPass string) bool {
    thisHash, err := hex.DecodeString(this.Pass);
    if (err != nil) {
        log.Warn().Err(err).Str("user", this.Email).Msg("Bad password hash for user.");
        return false;
    }

    salt, err := hex.DecodeString(this.Salt);
    if (err != nil) {
        log.Warn().Err(err).Str("user", this.Email).Msg("Bad salt for user.");
        return false;
    }

    otherHash := generateHash(hashPass, salt);

    return (subtle.ConstantTimeCompare(thisHash, otherHash) == 1);
}

// Merge another user's information into this user (email will not be merged).
// Empty values will not be merged.
// Returns true if any changes were made.
func (this *User) Merge(other *User) bool {
    changed := false;

    if ((other.Name != "") && (this.Name != other.Name)) {
        this.Name = other.Name;
        changed = true;
    }

    if ((other.Pass != "") && (this.Pass != other.Pass)) {
        this.Pass = other.Pass;
        this.Salt = other.Salt;
        changed = true;
    }

    if ((other.Role != RoleUnknown) && (this.Role != other.Role)) {
        this.Role = other.Role;
        changed = true;
    }

    if ((other.LMSID != "") && (this.LMSID != other.LMSID)) {
        this.LMSID = other.LMSID;
        changed = true;
    }

    return changed;
}

func generateHash(hashPass string, salt []byte) []byte {
    return argon2.IDKey([]byte(hashPass), salt, ARGON2_TIME, ARGON2_MEM_KB, ARGON2_THREADS, ARGON2_KEY_LEN_BYTES);
}

func SendUserAddEmail(user *User, pass string, generatedPass bool, userExists bool, dryRun bool, sleep bool) error {
    subject, body := composeUserAddEmail(user.Email, pass, generatedPass, userExists);

    if (dryRun) {
        log.Info().Str("email", user.Email).Msg("Doing a dry run, user will not be emailed.");
        log.Debug().Str("address", user.Email).Str("subject", subject).Str("body", body).Msg("Email not sent because of dry run.");
        return nil;
    }

    err := email.Send([]string{user.Email}, subject, body, false);
    if (err != nil) {
        log.Error().Err(err).Str("email", user.Email).Msg("Failed to send email.");
        return err;
    }

    log.Info().Str("email", user.Email).Msg("Registration email sent.");

    // Skip sleeping in testing mode.
    if (sleep && !config.TESTING_MODE.Get()) {
        time.Sleep(time.Duration(EMAIL_SLEEP_TIME));
    }

    return nil;
}

func composeUserAddEmail(address string, pass string, generatedPass bool, userExists bool) (string, string) {
    var subject string;
    var body string;

    if (userExists) {
        subject = "Autograder -- User Password Changed";
        body = fmt.Sprintf("Hello,\n\nThe password for '%s' has been changed.\n", address);
    } else {
        subject = "Autograder -- User Account Created";
        body = fmt.Sprintf("Hello,\n\nAn account with the username/email '%s' has been created.\n", address);
    }

    if (generatedPass) {
        body += fmt.Sprintf("The new password is '%s' (no quotes).\n", pass);
    }

    return subject, body;
}

func ToRowHeaader(delim string) string {
    parts := []string{"email", "name", "role", "lms-id"};
    return strings.Join(parts, delim);
}

func (this *User) ToRow(delim string) string {
    parts := []string{this.Email, this.Name, this.Role.String(), this.LMSID};
    return strings.Join(parts, delim);
}

func UsersPointerEqual(a []*User, b []*User) bool {
    if (len(a) != len(b)) {
        return false;
    }

    slices.SortFunc(a, UserPointerCompare);
    slices.SortFunc(b, UserPointerCompare);

    return slices.EqualFunc(a, b, UserPointerEqual);
}

func UserPointerEqual(a *User, b *User) bool {
    if (a == b) {
        return true;
    }

    if ((a == nil) || (b == nil)) {
        return false;
    }

    return (a.Email == b.Email) &&
            (a.Name == b.Name) &&
            (a.Role == b.Role) &&
            (a.LMSID == b.LMSID);
}

func UserPointerCompare(a *User, b *User) int {
    if (a == b) {
        return 0;
    }

    if (a == nil) {
        return 1;
    }

    if (b == nil) {
        return -1;
    }

    return strings.Compare(a.Email, b.Email);
}
