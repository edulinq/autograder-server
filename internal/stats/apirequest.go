package stats

import (
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

const (
	SENDER   = "sender"
	ENDPOINT = "endpoint"
	LOCATOR  = "locator"
	DURATION = "duration"
)

func AsyncStoreAPIRequestMetric(startTime timestamp.Timestamp, endTime timestamp.Timestamp, sender string, endpoint string, userEmail string, courseID string, assignmentID string, locator string) {
	attributes := map[string]any{
		SENDER:   sender,
		ENDPOINT: endpoint,
		DURATION: uint64((endTime - startTime).ToMSecs()),
	}

	if userEmail != "" {
		attributes[USER_EMAIL] = userEmail
	}

	if courseID != "" {
		attributes[COURSE_ID] = courseID
	}

	if assignmentID != "" {
		attributes[ASSIGNMENT_ID] = assignmentID
	}

	if locator != "" {
		attributes[LOCATOR] = locator
	}

	metric := &BaseMetric{
		Timestamp:  startTime,
		Attributes: attributes,
	}

	storeFunc := func() {
		err := StoreAPIRequestMetric(metric)
		if err != nil {
			log.Error("Failed to log API request metric.", err, metric)
		}
	}

	if config.UNIT_TESTING_MODE.Get() {
		storeFunc()
	} else {
		go storeFunc()
	}
}
