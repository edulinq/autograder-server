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

	AggregationQuery

	IncludeAPIRequestMetricField
	ExcludeAPIRequestMetricField
}

type IncludeAPIRequestMetricField struct {
	Sender       string `json:"include-sender,omitempty"`
	Endpoint     string `json:"include-endpoint,omitempty"`
	UserEmail    string `json:"include-user,omitempty"`
	CourseID     string `json:"include-course,omitempty"`
	AssignmentID string `json:"include-assignment,omitempty"`
	Locator      string `json:"include-locator,omitempty"`
}

type ExcludeAPIRequestMetricField struct {
	Sender       string `json:"exclude-sender,omitempty"`
	Endpoint     string `json:"exclude-endpoint,omitempty"`
	UserEmail    string `json:"exclude-user,omitempty"`
	CourseID     string `json:"exclude-course,omitempty"`
	AssignmentID string `json:"exclude-assignment,omitempty"`
	Locator      string `json:"exclude-locator,omitempty"`
}

func (this APIRequestMetricQuery) Match(record *APIRequestMetric) bool {
	if record == nil {
		return false
	}

	if !this.BaseQuery.Match(record) {
		return false
	}

	include := this.IncludeAPIRequestMetricField
	if (include.Sender != "") && (include.Sender != record.Sender) {
		return false
	}

	if (include.Endpoint != "") && (include.Endpoint != record.Endpoint) {
		return false
	}

	if (include.UserEmail != "") && (include.UserEmail != record.UserEmail) {
		return false
	}

	if (include.AssignmentID != "") && (include.AssignmentID != record.AssignmentID) {
		return false
	}

	if (include.CourseID != "") && (include.CourseID != record.CourseID) {
		return false
	}

	if (include.Locator != "") && (include.Locator != record.Locator) {
		return false
	}

	exclude := this.ExcludeAPIRequestMetricField
	if (exclude.Sender != "") && (exclude.Sender == record.Sender) {
		return false
	}

	if (exclude.Endpoint != "") && (exclude.Endpoint == record.Endpoint) {
		return false
	}

	if (exclude.UserEmail != "") && (exclude.UserEmail == record.UserEmail) {
		return false
	}

	if (exclude.AssignmentID != "") && (exclude.AssignmentID == record.AssignmentID) {
		return false
	}

	if (exclude.CourseID != "") && (exclude.CourseID == record.CourseID) {
		return false
	}

	if (exclude.Locator != "") && (exclude.Locator == record.Locator) {
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
