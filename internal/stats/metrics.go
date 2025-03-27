package stats

import (
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

type MetricAttribute string
type MetricType string

// Keys for the attributes field inside of Metric and Query.
const (
	// Common metric attribute keys
	DURATION_KEY MetricAttribute = "duration"
	ENDPOINT_KEY MetricAttribute = "endpoint"
	LOCATOR_KEY  MetricAttribute = "locator"
	SENDER_KEY   MetricAttribute = "sender"
	VALUE_KEY    MetricAttribute = "value"

	// Course-specific attribute keys
	ASSIGNMENT_ID_KEY MetricAttribute = "assignment"
	COURSE_ID_KEY     MetricAttribute = "course"
	USER_EMAIL_KEY    MetricAttribute = "user"

	// Specialized attribute keys
	ANALYSIS_KEY  MetricAttribute = "analysis-type"
	TASK_TYPE_KEY MetricAttribute = "task-type"
)

// Values for the type field inside of Metric and Query.
const (
	API_REQUEST_STATS_TYPE        MetricType = "api-request-stats"
	CODE_ANALYSIS_TIME_STATS_TYPE MetricType = "code-analysis-time-stats"
	GRADING_TIME_STATS_TYPE       MetricType = "grading-time-stats"
	TASK_TIME_STATS_TYPE          MetricType = "task-time-stats"
)

const ATTRIBUTES_KEY = "attributes"

type BaseMetric interface {
	GetTimestamp() timestamp.Timestamp
}

type Metric struct {
	Timestamp timestamp.Timestamp `json:"timestamp"`

	Type MetricType `json:"type"`

	// Additional attributes that are not standard enough to be formalized in fields.
	Attributes map[MetricAttribute]any `json:"attributes,omitempty"`
}

func (this Metric) GetTimestamp() timestamp.Timestamp {
	return this.Timestamp
}

type SystemMetrics struct {
	Metric

	CPUPercent       float64 `json:"cpu-percent"`
	MemPercent       float64 `json:"mem-percent"`
	NetBytesSent     uint64  `json:"net-bytes-sent"`
	NetBytesReceived uint64  `json:"net-bytes-received"`
}

func InsertIntoMapIfPresent(attributes map[MetricAttribute]any, key MetricAttribute, value any) {
	switch v := value.(type) {
	case string:
		if v != "" {
			attributes[key] = value
		}
	case nil:
		return
	default:
		attributes[key] = value
	}
}

func (this *Metric) LogValue() []*log.Attr {
	attrs := []*log.Attr{
		log.NewAttr("metric-type", this.Type),
	}

	courseID, ok := this.Attributes[COURSE_ID_KEY].(string)
	if ok {
		attrs = append(attrs, log.NewCourseAttr(courseID))
	}

	assignmentID, ok := this.Attributes[ASSIGNMENT_ID_KEY].(string)
	if ok {
		attrs = append(attrs, log.NewAssignmentAttr(assignmentID))
	}

	userEmail, ok := this.Attributes[USER_EMAIL_KEY].(string)
	if ok {
		attrs = append(attrs, log.NewUserAttr(userEmail))
	}

	return attrs
}

func AsyncStoreMetric(metric *Metric) {
	if metric == nil {
		return
	}

	storeFunc := func() {
		err := StoreMetric(metric)
		if err != nil {
			log.Error("Failed to log metric.", metric)
			return
		}
	}

	if config.UNIT_TESTING_MODE.Get() {
		storeFunc()
	} else {
		go storeFunc()
	}
}

func AsyncStoreCourseMetric(metric *Metric) {
	courseID, ok := metric.Attributes[COURSE_ID_KEY]
	if !ok || courseID == "" {
		log.Error("Cannot log course statistic without course ID.", metric)
		return
	}

	AsyncStoreMetric(metric)
}
