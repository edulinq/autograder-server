package logs

import (
	"fmt"

	"github.com/edulinq/autograder/internal/db"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/model"
)

// Query for log records.
// There are several steps before querying the database (parsing, validation, permissions check, etc),
// and this function does them all.
// The returned model.LocatableError should be shown to the caller (i.e. the user caused the error),
// while the returned error is a system error.
func Query(rawQuery log.RawLogQuery, contextUser *model.ServerUser) ([]*log.Record, *model.LocatableError, error) {
	parsedQuery, err := rawQuery.ParseJoin()
	if err != nil {
		message := fmt.Sprintf("Failed to parse log query: '%s'.", err.Error())
		locatableErr := &model.LocatableError{
			Locator:         "-1100",
			HideLocator:     false,
			InternalMessage: message,
			ExternalMessage: message,
		}

		return nil, locatableErr, nil
	}

	err = checkPermissions(parsedQuery, contextUser)
	if err != nil {
		locatableErr := &model.LocatableError{
			Locator:         "-1101",
			HideLocator:     true,
			InternalMessage: fmt.Sprintf("Bad permissions for loq query: '%s'.", err.Error()),
			ExternalMessage: "You do not have the correct permissions to execute this log query.",
		}

		return nil, locatableErr, nil
	}

	err = validateQuery(parsedQuery)
	if err != nil {
		// Permissions have already been checked, there is no risk of leaking information.
		message := fmt.Sprintf("Failed to validate log query: '%s'.", err.Error())
		locatableErr := &model.LocatableError{
			Locator:         "-1102",
			HideLocator:     false,
			InternalMessage: message,
			ExternalMessage: message,
		}

		return nil, locatableErr, nil
	}

	records, err := db.GetLogRecords(*parsedQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to execute log query: '%w'.", err)
	}

	return records, nil, nil
}
