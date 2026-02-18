package log

import "fmt"

type LongLogString string

const DEFAULT_MAX_LOG_STRING_LENGTH = 1000

func (s LongLogString) String() string {
	str := string(s)
	if len(str) > DEFAULT_MAX_LOG_STRING_LENGTH {
		return fmt.Sprintf("%s... [truncated %d more bytes]",
			str[:DEFAULT_MAX_LOG_STRING_LENGTH],
			len(str)-DEFAULT_MAX_LOG_STRING_LENGTH)
	}
	return str
}
