package model

import (
    "fmt"
    "strings"

	"github.com/edulinq/autograder/internal/model"
)

const USER_REFERENCE_DELIM = ""

type ServerUserReference string

// TODO: think about avoiding root, making sure the delim only happens once (be careful of false positives on emails).
// Maybe check if it's an email first, then move on to other checks?
func (this *ServerUserReference) Validate() error {
    if this == nil {
        // TODO: Improve error message.
        return fmt.Errorf("Cannot have a nil user reference.")
    }

    userReference := strings.ToLower(strings.TrimSpace(*this))

    if strings.Contains(userReference, "@") {
        return nil
    }

    userReference = strings.TrimPrefix(userReference, "-")

    if userReference == "root" {
        return fmt.Errorf("A server user reference cannot reference the root user.")
    }

    if userReference == "*" {
        return nil
    }

    if GetServerUserRole[userReference] != ServerRoleUnknown {
        return nil
    }

    courses := GetCourses()
    _, ok := courses[userReference]
    if ok {
        return nil
    }

    referenceParts := strings.Split(this, USER_REFERENCE_DELIM)
    if len(referenceParts) != 2 {
        return fmt.Errorf("A non-email user reference must have two parts.")
    }
}
