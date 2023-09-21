package model

import (
    "crypto/subtle"
    "encoding/hex"
    "fmt"

    "github.com/rs/zerolog/log"
    "golang.org/x/crypto/argon2"

    "github.com/eriq-augustine/autograder/util"
)

const (
    SALT_LENGTH_BYTES = 16;

    ARGON2_KEY_LEN_BYTES = 32;
    ARGON2_MEM_KB = 64 * 1024;
    ARGON2_THREADS = 1;
    ARGON2_TIME = 1;
)

type User struct {
    Email string `json:"email"`
    DisplayName string `json:"display-name"`
    Role UserRole `json:"role"`
    Pass string `json:"pass"`
    Salt string `json:"salt"`
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
