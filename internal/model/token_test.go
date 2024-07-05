package model

import (
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestTokenBase(test *testing.T) {
	pass := util.Sha256HexFromString("foo")
	badPass := util.Sha256HexFromString("bar")

	salt, err := NewRandomSalt()
	if err != nil {
		test.Fatalf("Failed to generate salt: '%v'.", err)
	}

	token, err := NewToken(pass, salt, TokenSourceServer, "")
	if err != nil {
		test.Fatalf("Failed to generate token: '%v'.", err)
	}

	match, err := token.Check(pass, salt)
	if err != nil {
		test.Fatalf("Failed to check good match: '%v'.", err)
	}

	if !match {
		test.Fatalf("Token did not match when it should have.")
	}

	match, err = token.Check(badPass, salt)
	if err != nil {
		test.Fatalf("Failed to check bad match: '%v'.", err)
	}

	if match {
		test.Fatalf("Token did match when it should not have.")
	}
}

func TestTokenRandom(test *testing.T) {
	salt, err := NewRandomSalt()
	if err != nil {
		test.Fatalf("Failed to generate salt: '%v'.", err)
	}

	clearPass, token, err := NewRandomToken(salt, TokenSourceServer, "")
	if err != nil {
		test.Fatalf("Failed to generate random token: '%v'.", err)
	}

	pass := util.Sha256HexFromString(clearPass)
	badPass := util.Sha256HexFromString("bar")

	match, err := token.Check(pass, salt)
	if err != nil {
		test.Fatalf("Failed to check good match: '%v'.", err)
	}

	if !match {
		test.Fatalf("Token did not match when it should have.")
	}

	match, err = token.Check(badPass, salt)
	if err != nil {
		test.Fatalf("Failed to check bad match: '%v'.", err)
	}

	if match {
		test.Fatalf("Token did match when it should not have.")
	}
}
