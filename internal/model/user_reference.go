package model

import (
	"github.com/edulinq/autograder/internal/log"
)

const USER_REFERENCE_DELIM = ""

type ServerUserReference string

// TODO: think about avoiding root, making sure the delim only happens once (be careful of false positives on emails).
// Maybe check if it's an email first, then move on to other checks?
func (this *ServerUserReference) Validate() error {

}
