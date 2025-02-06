package stats

import (
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
	"github.com/edulinq/autograder/internal/util"
)

type CourseMetricType string

const (
	CourseMetricTypeUnknown          CourseMetricType = ""
	CourseMetricTypeGradingTime                       = "grading-time"
	CourseMetricTypeTaskTime                          = "task-time"
	CourseMetricTypeCodeAnalysisTime                  = "code-analysis-time"
)

const (
	ATTRIBUTE_KEY_TASK     = "task-name"
	ATTRIBUTE_KEY_ANALYSIS = "analysis-type"
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
	CourseMetricTypeUnknown:          string(CourseMetricTypeUnknown),
	CourseMetricTypeGradingTime:      string(CourseMetricTypeGradingTime),
	CourseMetricTypeCodeAnalysisTime: string(CourseMetricTypeCodeAnalysisTime),
}

var stringToCourseMetricType = map[string]CourseMetricType{
	string(CourseMetricTypeUnknown):          CourseMetricTypeUnknown,
	string(CourseMetricTypeGradingTime):      CourseMetricTypeGradingTime,
	string(CourseMetricTypeCodeAnalysisTime): CourseMetricTypeCodeAnalysisTime,
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

func (this *CourseMetric) LogValue() []*log.Attr {
	attrs := []*log.Attr{
		log.NewAttr("course-metric-type", this.Type),
	}

	if this.CourseID != "" {
		attrs = append(attrs, log.NewCourseAttr(this.CourseID))
	}

	if this.AssignmentID != "" {
		attrs = append(attrs, log.NewAssignmentAttr(this.AssignmentID))
	}

	if this.UserEmail != "" {
		attrs = append(attrs, log.NewUserAttr(this.UserEmail))
	}

	return attrs
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

// Store a course metric without blocking (unless this is running in test mode, then it will block).
// Course ID is required, and all provided IDs should already be validated.
func AsyncStoreCourseMetric(metric *CourseMetric) {
	if metric == nil {
		return
	}

	if metric.CourseID == "" {
		log.Error("Cannot log course statistic without course ID.", metric)
		return
	}

	storeFunc := func() {
		err := StoreCourseMetric(metric)
		if err != nil {
			log.Error("Failed to log course metric.", err, metric)
		}
	}

	if config.UNIT_TESTING_MODE.Get() {
		storeFunc()
	} else {
		go storeFunc()
	}
}

func AsyncStoreCourseGradingTime(startTime timestamp.Timestamp, endTime timestamp.Timestamp, courseID string, assignmentID string, userEmail string) {
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

	AsyncStoreCourseMetric(metric)
}

func AsyncStoreCourseTaskTime(startTime timestamp.Timestamp, endTime timestamp.Timestamp, courseID string, assignmentID string, userEmail string, taskName string) {
	metric := &CourseMetric{
		BaseMetric: BaseMetric{
			Timestamp:  startTime,
			Attributes: map[string]any{ATTRIBUTE_KEY_TASK: taskName},
		},
		Type:         CourseMetricTypeTaskTime,
		CourseID:     courseID,
		AssignmentID: assignmentID,
		UserEmail:    userEmail,
		Value:        uint64((endTime - startTime).ToMSecs()),
	}

	AsyncStoreCourseMetric(metric)
}
