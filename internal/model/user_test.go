package model

import (
    "testing"
)

var testPasswords []string = []string{
    "a",
    "A",
    "1",
    "-1",
    "aaa",
    "AAA",
    "123",
}

func TestUserPasswordTable(test *testing.T) {
    for i, a := range testPasswords {
        user := &User{};
        user.SetPassword(a);

        for j, b := range testPasswords {
            equals := user.CheckPassword(b);
            if ((i == j) && !equals) {
                test.Errorf("Passwords hashes are not equal when they should be: '%s' and '%s'.", a, b);
            } else if ((i != j) && equals) {
                test.Errorf("Passwords hashes are equal when they should not be: '%s' and '%s'.", a, b);
            }
        }
    }
}
