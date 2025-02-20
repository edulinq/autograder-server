package stats

import (
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

type RequestMetric struct {
	BaseMetric

	Endpoint     string `json:"endpoint"`
	Locator      string `json:"locator,omitempty"`
	CourseID     string `json:"course,omitempty"`
	AssignmentID string `json:"assignment,omitempty"`
	UserEmail    string `json:"user,omitempty"`
	IPAddress    string `json:ip-address`
	Value        uint64 `json:"duration"`
}

func AsyncStoreRequestMetric(startTime timestamp.Timestamp, endTime timestamp.Timestamp, courseID string, assignmentID string, userEmail string, endpoint string, locator string, ipAddress string) {
	metric := &RequestMetric{
		BaseMetric: BaseMetric{
			Timestamp: startTime,
		},
		Endpoint:     endpoint,
		Locator:      locator,
		CourseID:     courseID,
		AssignmentID: assignmentID,
		UserEmail:    userEmail,
		IPAddress:    ipAddress,
		Value:        uint64((endTime - startTime).ToMSecs()),
	}

	storeFunc := func() {
		err := StoreRequestMetric(metric)
		if err != nil {
			log.Error("Failed to log request metric.", err, metric)
		}
	}

	if config.UNIT_TESTING_MODE.Get() {
		storeFunc()
	} else {
		go storeFunc()
	}
}
