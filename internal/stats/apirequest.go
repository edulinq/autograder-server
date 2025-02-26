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

type APIRequestMetricQuery struct {
	BaseQuery

	Sender       string `json:"target-sender"`
	Endpoint     string `json:"target-endpoint"`
	UserEmail    string `json:"target-user,omitempty"`
	CourseID     string `json:"target-course,omitempty"`
	AssignmentID string `json:"target-assignment,omitempty"`
	Locator      string `json:"target-locator"`
}

func (this APIRequestMetricQuery) Match(record *APIRequestMetric) bool {
	if record == nil {
		return false
	}

	if !this.BaseQuery.Match(record) {
		return false
	}

	if (this.Sender != "") && (this.Sender != record.Sender) {
		return false
	}

	if (this.Endpoint != "") && (this.Endpoint != record.Endpoint) {
		return false
	}

	if (this.UserEmail != "") && (this.UserEmail != record.UserEmail) {
		return false
	}

	if (this.AssignmentID != "") && (this.AssignmentID != record.AssignmentID) {
		return false
	}

	if (this.CourseID != "") && (this.CourseID != record.CourseID) {
		return false
	}

	if (this.Locator != "") && (this.Locator != record.Locator) {
		return false
	}

	return true
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
