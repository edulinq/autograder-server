package logs

import (
	"fmt"

	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// Query for log records.
// There are several steps before querying the database (parsing, validation, permissions check, etc),
// and this function does them all.
func Query(rawQuery log.RawLogQuery, contextUser *model.ServerUser) ([]*log.Record, *LocatableError) {
	parsedQuery, err := rawQuery.ParseJoin()
	if err != nil {
		message := fmt.Sprintf("Failed to parse log query: '%s'.", err.Error())
		return nil, &LocatableError{
			Locator:         "-30001",
			HideLocator:     false,
			InternalMessage: message,
			ExternalMessage: message,
		}
	}

	err = checkPermissions(parsedQuery, contextUser)
	if err != nil {
		return nil, &LocatableError{
			Locator:         "-30002",
			HideLocator:     true,
			InternalMessage: fmt.Sprintf("Bad permissions for loq query: '%s'.", err.Error()),
			ExternalMessage: "You do not have the correct permissions to execute this log query.",
		}
	}

	err = validateQuery(&parsedQuery)
	if err != nil {
		// Permissions have already been checked, there is no risk of leaking information.
		message := fmt.Sprintf("Failed to validate log query: '%s'.", err.Error())
		return nil, &LocatableError{
			Locator:         "-30003",
			HideLocator:     false,
			InternalMessage: message,
			ExternalMessage: message,
		}
	}

	records, err := db.GetLogRecords(parsedQuery)
	if err != nil {
		message := fmt.Sprintf("Failed to execute log query: '%s'.", err.Error())
		return nil, &LocatableError{
			Locator:         "-30004",
			HideLocator:     false,
			InternalMessage: message,
			ExternalMessage: message,
		}
	}

	return records
}
