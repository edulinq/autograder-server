package stats

import (
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

const (
	DURATION_KEY = "duration"
	ENDPOINT_KEY = "endpoint"
	LOCATOR_KEY  = "locator"
	SENDER_KEY   = "sender"
)

func AsyncStoreAPIRequestMetric(startTime timestamp.Timestamp, endTime timestamp.Timestamp, sender string, endpoint string, userEmail string, courseID string, assignmentID string, locator string) {
	attributes := map[string]any{
		SENDER_KEY:   sender,
		ENDPOINT_KEY: endpoint,
		DURATION_KEY: uint64((endTime - startTime).ToMSecs()),
	}

	if userEmail != "" {
		attributes[USER_EMAIL_KEY] = userEmail
	}

	if courseID != "" {
		attributes[COURSE_ID_KEY] = courseID
	}

	if assignmentID != "" {
		attributes[ASSIGNMENT_ID_KEY] = assignmentID
	}

	if locator != "" {
		attributes[LOCATOR_KEY] = locator
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
