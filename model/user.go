package model

// It is expected that any cleartext password passed into functions here
// are already a hex encoding of a sha256 hash of the original cleartext
// see util.Sha256Hex().

import (
    "crypto/subtle"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/rs/zerolog/log"
    "golang.org/x/crypto/argon2"

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
    DisplayName string `json:"display-name"`
    Role UserRole `json:"role"`
    Pass string `json:"pass"`
    Salt string `json:"salt"`

    CanvasID string `json:"canvas-id,omitempty"`
}

// Sets the password and generates a new salt.
func (this *User) SetPassword(cleartext string) error {
    salt, err := util.RandBytes(SALT_LENGTH_BYTES);
    if (err != nil) {
        return fmt.Errorf("Could not generate salt: '%w'.", err);
    }

    pass := generateHash(cleartext, salt);

    this.Salt = hex.EncodeToString(salt);
    this.Pass = hex.EncodeToString(pass);

    return nil;
}

// Return true if the password matches the hash, false otherwise.
// Any errors (which can only come from bad hex strings) will be logged and ignored (false will be returned).
func (this *User) CheckPassword(cleartext string) bool {
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

    otherHash := generateHash(cleartext, salt);

    return (subtle.ConstantTimeCompare(thisHash, otherHash) == 1);
}

func generateHash(cleartext string, salt []byte) []byte {
    return argon2.IDKey([]byte(cleartext), salt, ARGON2_TIME, ARGON2_MEM_KB, ARGON2_THREADS, ARGON2_KEY_LEN_BYTES);
}

func LoadUsersFile(path string) (map[string]*User, error) {
    users := make(map[string]*User);

    if (!util.PathExists(path)) {
        return users, nil;
    }

    err := util.JSONFromFile(path, &users);
    if (err != nil) {
        return nil, err;
    }

    return users, nil;
}

func SaveUsersFile(path string, users map[string]*User) error {
    return util.ToJSONFileIndent(users, path);
}

// Return a user that is either new or a merged with the existing user (depending on force).
// If a user exists (and force is true), then the user will be updated.
// New users will just be retuturned and not be added to |users|.
func NewOrMergeUser(users map[string]*User, email string, name string, stringRole string, pass string, force bool) (*User, bool, error) {
    user := users[email];
    userExists := (user != nil);

    if (userExists && !force) {
        return nil, true, fmt.Errorf("User '%s' already exists, cannot add.", email);
    }

    if (!userExists) {
        user = &User{Email: email};
    }

    if (name != "") {
        user.DisplayName = name;
    } else  if (user.DisplayName == "") {
        user.DisplayName = email;
    }

    // Note the slightly tricky conditions here.
    // Only error if the string role is bad and there is not an existing good role.
    role := GetRole(stringRole);
    if (role != Unknown) {
        user.Role = role;
    } else if (user.Role == Unknown) {
        return nil, false, fmt.Errorf("Unknown role: '%s'.", stringRole);
    }

    hashPass := util.Sha256Hex([]byte(pass));

    err := user.SetPassword(hashPass);
    if (err != nil) {
        return nil, false, fmt.Errorf("Could not set password: '%w'.", err);
    }

    return user, userExists, nil;
}

func SendUserAddEmail(user *User, pass string, generatedPass bool, userExists bool, dryRun bool, sleep bool) {
    subject, body := composeUserAddEmail(user.Email, pass, generatedPass, userExists);

    if (dryRun) {
        fmt.Printf("Doing a dry run, user '%s' will not be emailed.\n", user.Email);
        log.Debug().Str("address", user.Email).Str("subject", subject).Str("body", body).Msg("Email not sent because of dry run.");
    } else {
        err := email.Send([]string{user.Email}, subject, body);
        if (err != nil) {
            log.Error().Err(err).Str("email", user.Email).Msg("Failed to send email.");
        } else {
            fmt.Printf("Registration email send to '%s'.\n", user.Email);
        }

        if (sleep) {
            time.Sleep(time.Duration(EMAIL_SLEEP_TIME));
        }
    }
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
