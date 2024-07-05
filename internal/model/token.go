package model

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/edulinq/autograder/internal/common"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/util"
)

// Token sources refer to where/how a token was created.
// E.g., tokens could be created by the request of the user, or by the server when adding a user.
type TokenSource string

const DEFAULT_SALT string = "c75385ab94f66b3454e93cb0d0546fc1"
const TOKEN_PASSWORD_NAME = "password"

const (
	TokenSourceUnknown  TokenSource = ""
	TokenSourceServer               = "server"
	TokenSourceUser                 = "user"
	TokenSourceAdmin                = "admin"
	TokenSourcePassword             = "password"
)

// Tokens refer to any hex string that is used for authentication.
// They can come from different sources but all act the same.
type Token struct {
	ID           string           `json:"id"`
	HexDigest    string           `json:"hex-digest"`
	Source       TokenSource      `json:"source"`
	Name         string           `json:"name"`
	CreationTime common.Timestamp `json:"creation-time"`
	AccessTime   common.Timestamp `json:"access-time"`
}

const (
	DEFAULT_TOKEN_LEN = 32
	SALT_LENGTH_BYTES = 32

	ARGON2_KEY_LEN_BYTES = 32
	ARGON2_MEM_KB        = 64 * 1024
	ARGON2_THREADS       = 4
	ARGON2_TIME          = 1
)

// Create a new token.
// Note that it is suggested (but not required) that the input be the hex encoding output from a Sha256 digest.
// This is not required, but RandomToken() will always take this step.
func NewToken(input string, salt string, source TokenSource, name string) (*Token, error) {
	if input == "" {
		return nil, fmt.Errorf("Token ('%s') has zero-length input.", name)
	}

	digestBytes, err := generateDigest(input, salt)
	if err != nil {
		return nil, fmt.Errorf("Could not generate digest for token ('%s'): '%w'.", name, err)
	}

	digest := hex.EncodeToString(digestBytes)

	now := common.NowTimestamp()

	token := &Token{
		ID:           util.UUID(),
		HexDigest:    digest,
		Source:       source,
		Name:         name,
		CreationTime: now,
		AccessTime:   now,
	}

	return token, token.Validate()
}

func MustNewToken(input string, salt string, source TokenSource, name string) *Token {
	token, err := NewToken(input, salt, source, name)
	if err != nil {
		log.Fatal("Failed to create new token.", err)
	}

	return token
}

// Generate a random input string, sha256 that random string, use the output to create a token, and return the random cleartext and token.
// Remember that the text before sha256 will be returned.
func NewRandomToken(salt string, source TokenSource, name string) (string, *Token, error) {
	cleartext, err := RandomCleartext()
	if err != nil {
		return "", nil, fmt.Errorf("Failed to generate random token ('%s'): '%w'.", name, err)
	}

	input := util.Sha256HexFromString(cleartext)

	token, err := NewToken(input, salt, source, name)
	if err != nil {
		return "", nil, err
	}

	return cleartext, token, nil
}

func MustNewRandomToken(salt string, source TokenSource, name string) (string, *Token) {
	cleartext, token, err := NewRandomToken(salt, source, name)
	if err != nil {
		log.Fatal("Failed to create new random token.", err)
	}

	return cleartext, token
}

func RandomCleartext() (string, error) {
	return util.RandHex(DEFAULT_TOKEN_LEN)
}

// Get a new random salt.
// Salts must be hex encoded strings.
// If a salt could not be generated, an error will be logged and a default salt will be returned.
func ShouldNewRandomSalt() string {
	salt, err := NewRandomSalt()
	if err != nil {
		log.Error("Failed to generate salt.", err)
		return DEFAULT_SALT
	}

	return salt
}

// Get a new random salt.
// Salts must be hex encoded strings.
func NewRandomSalt() (string, error) {
	saltBytes, err := util.RandBytes(SALT_LENGTH_BYTES)
	if err != nil {
		return "", fmt.Errorf("Failed to generate salt: '%w'.", err)
	}

	return hex.EncodeToString(saltBytes), nil
}

// Check if some input matches this token.
// As with NewToken(), the input is suggested (but not required) to the hex encoding of a Sha256 digest.
// The salt must be a hex encoded string.
// If the input matches, then true will be returned and the token's access time will be set,
// false will otherwise be returned.
func (this *Token) Check(input string, salt string) (bool, error) {
	now := common.NowTimestamp()

	thisDigestBytes, err := hex.DecodeString(this.HexDigest)
	if err != nil {
		return false, fmt.Errorf("Token ('%s') has a salt that is not valid hex.", this.Name)
	}

	otherDigestBytes, err := generateDigest(input, salt)
	if err != nil {
		return false, fmt.Errorf("Could not generate digest for comparison with token ('%s'): '%w'.", this.Name, err)
	}

	match := (subtle.ConstantTimeCompare(thisDigestBytes, otherDigestBytes) == 1)
	if match {
		this.AccessTime = now
	}

	return match, nil
}

func (this *Token) Validate() error {
	if this == nil {
		return fmt.Errorf("Token is nil.")
	}

	this.Name = strings.TrimSpace(this.Name)

	if this.HexDigest == "" {
		return fmt.Errorf("Token ('%s') has empty digest.", this.Name)
	}

	_, err := hex.DecodeString(this.HexDigest)
	if err != nil {
		return fmt.Errorf("Token ('%s') has a digest that is not valid hed: '%w'.", this.Name, err)
	}

	if this.Source == TokenSourceUnknown {
		return fmt.Errorf("Token ('%s') has unknown source.", this.Name)
	}

	return nil
}

func (this *Token) Clone() *Token {
	return &Token{
		ID:           this.ID,
		HexDigest:    this.HexDigest,
		Source:       this.Source,
		Name:         this.Name,
		CreationTime: this.CreationTime,
		AccessTime:   this.AccessTime,
	}
}

func TokenPointerEqual(a *Token, b *Token) bool {
	return TokenPointerCompare(a, b) == 0
}

// Note that the token's ID is not used in comparison checks,
// i.e., tokens with different IDs can still be considered equal.
func TokenPointerCompare(a *Token, b *Token) int {
	if (a == nil) && (b == nil) {
		return 0
	}

	if a == nil {
		return -1
	}

	if b == nil {
		return 1
	}

	result := strings.Compare(string(a.Source), string(b.Source))
	if result != 0 {
		return result
	}

	result = strings.Compare(string(a.Name), string(b.Name))
	if result != 0 {
		return result
	}

	result = strings.Compare(string(a.HexDigest), string(b.HexDigest))
	if result != 0 {
		return result
	}

	return 0
}

// Generate a digest (bytes) for an input and salt.
// The salt must be a hex encoded string.
func generateDigest(input string, salt string) ([]byte, error) {
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return nil, fmt.Errorf("Salt is not valid hex: '%w'.", err)
	}

	return argon2.IDKey([]byte(input), saltBytes, ARGON2_TIME, ARGON2_MEM_KB, ARGON2_THREADS, ARGON2_KEY_LEN_BYTES), err
}
