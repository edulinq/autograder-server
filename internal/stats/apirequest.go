package stats

import (
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

type APIRequestMetric struct {
	BaseMetric

	Sender       string `json:"sender"`
	Endpoint     string `json:"endpoint"`
	UserEmail    string `json:"user,omitempty"`
	AssignmentID string `json:"assignment,omitempty"`
	CourseID     string `json:"course,omitempty"`
	Locator      string `json:"locator,omitempty"`

	Duration uint64 `json:"duration"`
}

func AsyncStoreAPIRequestMetric(startTime timestamp.Timestamp, endTime timestamp.Timestamp, sender string, endpoint string, userEmail string, courseID string, assignmentID string, locator string) {
	metric := &APIRequestMetric{
		BaseMetric: BaseMetric{
			Timestamp: startTime,
		},
		Sender:       sender,
		Endpoint:     endpoint,
		UserEmail:    userEmail,
		CourseID:     courseID,
		AssignmentID: assignmentID,
		Locator:      locator,
		Duration:     uint64((endTime - startTime).ToMSecs()),
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
