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
	Duration     uint64 `json:"duration"`
	UserEmail    string `json:"user,omitempty"`
	CourseID     string `json:"course,omitempty"`
	AssignmentID string `json:"assignment,omitempty"`
	Locator      string `json:"locator,omitempty"`
}

type APIRequestMetricQuery struct {
	BaseMetric
}

func AsyncStoreAPIRequestMetric(startTime timestamp.Timestamp, endTime timestamp.Timestamp, sender string, endpoint string, userEmail string, courseID string, assignmentID string, locator string) {
	metric := &APIRequestMetric{
		BaseMetric: BaseMetric{
			Timestamp: startTime,
		},
		Endpoint:     endpoint,
		Locator:      locator,
		CourseID:     courseID,
		AssignmentID: assignmentID,
		UserEmail:    userEmail,
		Sender:       sender,
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
