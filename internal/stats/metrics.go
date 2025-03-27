package stats

import (
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

const (
	ATTRIBUTES_KEY               = "attributes"
	TYPE_KEY                     = "type"
	DURATION_KEY                 = "duration"
	ENDPOINT_KEY                 = "endpoint"
	LOCATOR_KEY                  = "locator"
	SENDER_KEY                   = "sender"
	API_REQUEST_STATS_KEY        = "api-request-stats"
	GRADING_TIME_STATS_KEY       = "grading-time-stats"
	TASK_TIME_STATS_KEY          = "task-time-stats"
	CODE_ANALYSIS_TIME_STATS_KEY = "code-analysis-time-stats"
	ATTRIBUTE_KEY_TASK           = "task-type"
	ATTRIBUTE_KEY_ANALYSIS       = "analysis-type"
	ASSIGNMENT_ID_KEY            = "assignment"
	COURSE_ID_KEY                = "course"
	TASK_TYPE_KEY                = "task-type"
	COURSE_TYPE_KEY              = "type"
	VALUE_KEY                    = "value"
	USER_EMAIL_KEY               = "user"
)

type Metric interface {
	GetTimestamp() timestamp.Timestamp
}

type BaseMetric struct {
	Timestamp timestamp.Timestamp `json:"timestamp"`

	Type string `json:"type"`

	// Additional attributes that are not standard enough to be formalized in fields.
	Attributes map[string]any `json:"attributes,omitempty"`
}

func (this BaseMetric) GetTimestamp() timestamp.Timestamp {
	return this.Timestamp
}

type SystemMetrics struct {
	BaseMetric

	CPUPercent       float64 `json:"cpu-percent"`
	MemPercent       float64 `json:"mem-percent"`
	NetBytesSent     uint64  `json:"net-bytes-sent"`
	NetBytesReceived uint64  `json:"net-bytes-received"`
}

func AddIfNotEmpty(attributes map[string]any, key string, value any) {
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

func (this *BaseMetric) LogValue() []*log.Attr {
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

func AsyncStoreMetric(metric *BaseMetric) {
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
