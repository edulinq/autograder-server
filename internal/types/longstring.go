package types

import "github.com/edulinq/autograder/internal/util"

type LongString string

const MAX_LOG_STRING_LENGTH = 1000

func (this LongString) String() string {

	return util.ClipStringWithHash(string(this), MAX_LOG_STRING_LENGTH)
}
