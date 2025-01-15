package stats

import (
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type CourseMetricType string

const (
	CourseMetricTypeUnknown     CourseMetricType = ""
	CourseMetricTypeGradingTime                  = "grading-time"
)

type CourseMetric struct {
	BaseMetric

	Type CourseMetricType `json:"type"`

	CourseID     string `json:"course,omitempty"`
	AssignmentID string `json:"assignment,omitempty"`
	UserEmail    string `json:"user,omitempty"`

	Value uint64 `json:"duration"`
}

type CourseMetricQuery struct {
	BaseQuery

	Type CourseMetricType `json:"target-type"`

	CourseID     string `json:"target-course,omitempty"`
	AssignmentID string `json:"target-assignment,omitempty"`
	UserEmail    string `json:"target-user,omitempty"`
}

var courseMetricTypeToString = map[CourseMetricType]string{
	CourseMetricTypeUnknown:     string(CourseMetricTypeUnknown),
	CourseMetricTypeGradingTime: string(CourseMetricTypeGradingTime),
}

var stringToCourseMetricType = map[string]CourseMetricType{
	string(CourseMetricTypeUnknown):     CourseMetricTypeUnknown,
	string(CourseMetricTypeGradingTime): CourseMetricTypeGradingTime,
}

func (this CourseMetricType) MarshalJSON() ([]byte, error) {
	return util.MarshalEnum(this, courseMetricTypeToString)
}

func (this *CourseMetricType) UnmarshalJSON(data []byte) error {
	value, err := util.UnmarshalEnum(data, stringToCourseMetricType, true)
	if err == nil {
		*this = *value
	}

	return err
}

func (this CourseMetricQuery) Match(record *CourseMetric) bool {
	if record == nil {
		return false
	}

	if !this.BaseQuery.Match(record) {
		return false
	}

	if (this.Type != CourseMetricTypeUnknown) && (this.Type != record.Type) {
		return false
	}

	if (this.CourseID != "") && (this.CourseID != record.CourseID) {
		return false
	}

	if (this.AssignmentID != "") && (this.AssignmentID != record.AssignmentID) {
		return false
	}

	if (this.UserEmail != "") && (this.UserEmail != record.UserEmail) {
		return false
	}

	return true
}

// Store a grading time metric without blocking (unless this is running in test mode, then it will block).
// Course ID is required, and all provided IDs should already be validated.
func AsyncStoreCourseGradingTime(startTime timestamp.Timestamp, endTime timestamp.Timestamp, courseID string, assignmentID string, userEmail string) {
	storeFunc := func() {
		err := storeCourseGradingTime(startTime, endTime, courseID, assignmentID, userEmail)
		if err != nil {
			log.Error("Failed to log course grading time.", err, log.NewCourseAttr(courseID), log.NewAssignmentAttr(assignmentID), log.NewUserAttr(userEmail))
		}
	}

	if config.UNIT_TESTING_MODE.Get() {
		storeFunc()
	} else {
		go storeFunc()
	}
}

func storeCourseGradingTime(startTime timestamp.Timestamp, endTime timestamp.Timestamp, courseID string, assignmentID string, userEmail string) error {
	if courseID == "" {
		return fmt.Errorf("Cannot log course statistic without course ID.")
	}

	metric := &CourseMetric{
		BaseMetric: BaseMetric{
			Timestamp: startTime,
		},
		Type:         CourseMetricTypeGradingTime,
		CourseID:     courseID,
		AssignmentID: assignmentID,
		UserEmail:    userEmail,
		Value:        uint64((endTime - startTime).ToMSecs()),
	}

	return StoreCourseMetric(metric)
}
