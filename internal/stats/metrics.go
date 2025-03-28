package stats

import (
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

type MetricAttribute string
type MetricType string
type MetricTypeInfo struct {
	RequiresCourseID bool
}

// Keys for the attributes field inside of Metric and Query.
const (
	UNKNOWN_METRIC_ATTRIBUTE_KEY MetricAttribute = ""
	ASSIGNMENT_ID_KEY                            = "assignment"
	ANALYSIS_KEY                                 = "analysis-type"
	COURSE_ID_KEY                                = "course"
	ENDPOINT_KEY                                 = "endpoint"
	LOCATOR_KEY                                  = "locator"
	SENDER_KEY                                   = "sender"
	TASK_TYPE_KEY                                = "task-type"
	USER_EMAIL_KEY                               = "user"
)

// Values for the type field inside of Metric and Query.
const (
	UNKNOWN_METRIC_ATTRIBUTE_TYPE MetricType = ""
	API_REQUEST_STATS_TYPE                   = "api-request-stats"
	CODE_ANALYSIS_TIME_STATS_TYPE            = "code-analysis-time-stats"
	GRADING_TIME_STATS_TYPE                  = "grading-time-stats"
	TASK_TIME_STATS_TYPE                     = "task-time-stats"
)

const ATTRIBUTES_KEY = "attributes"

var metricTypeRequiresCourse = map[MetricType]MetricTypeInfo{
	API_REQUEST_STATS_TYPE:        {RequiresCourseID: false},
	CODE_ANALYSIS_TIME_STATS_TYPE: {RequiresCourseID: true},
	GRADING_TIME_STATS_TYPE:       {RequiresCourseID: true},
	TASK_TIME_STATS_TYPE:          {RequiresCourseID: true},
}

type TimestampedMetric interface {
	GetTimestamp() timestamp.Timestamp
}

type Metric struct {
	Timestamp timestamp.Timestamp `json:"timestamp"`

	Type MetricType `json:"type"`

	Value float64 `json:"value"`

	// Additional attributes that are not standard enough to be formalized in fields.
	Attributes map[MetricAttribute]any `json:"attributes,omitempty"`
}

func (this Metric) GetTimestamp() timestamp.Timestamp {
	return this.Timestamp
}

func (this *Metric) LogValue() []*log.Attr {
	attrs := []*log.Attr{
		log.NewAttr("metric-type", this.Type),
	}

	attrs = append(attrs, log.NewAttr("metric-attributes", this.Attributes))

	return attrs
}

// Ensure a course ID is provided when required by the metric type.
func (this Metric) hasRequiredCourseID() bool {
	requiresCourseID := metricTypeRequiresCourse[this.Type].RequiresCourseID

	courseID, ok := this.Attributes[COURSE_ID_KEY]
	if requiresCourseID && (!ok || courseID == nil || courseID == "") {
		return false
	}

	return true
}

func (this *Metric) Validate() error {
	if this == nil {
		return fmt.Errorf("No metric was given.")
	}

	if this.Type == UNKNOWN_METRIC_ATTRIBUTE_TYPE {
		return fmt.Errorf("Metric attribute was not set.")
	}

	if this.Timestamp.IsZero() {
		return fmt.Errorf("Metric timestamp was not set.")
	}

	if !this.hasRequiredCourseID() {
		return fmt.Errorf("Metric type '%s' requires a course ID", this.Type)
	}

	for field, value := range this.Attributes {
		if field == UNKNOWN_METRIC_ATTRIBUTE_KEY {
			return fmt.Errorf("Metric attribute field was empty.")
		}

		if value == nil || value == "" {
			return fmt.Errorf("Metric attribute value was empty.")
		}
	}

	return nil
}

func AsyncStoreMetric(metric *Metric) {
	err := metric.Validate()
	if err != nil {
		log.Error("Failed to validate metric.", err)
		return
	}

	storeFunc := func() {
		err := StoreMetric(metric)
		if err != nil {
			log.Error("Failed to store metric.", metric)
			return
		}
	}

	if config.UNIT_TESTING_MODE.Get() {
		storeFunc()
	} else {
		go storeFunc()
	}
}

func (this *Metric) SetAssignmentID(id string) {
	if id != "" {
		this.Attributes[ASSIGNMENT_ID_KEY] = id
	}
}

func (this *Metric) SetAnalysisType(analysis string) {
	if analysis != "" {
		this.Attributes[ANALYSIS_KEY] = analysis
	}
}

func (this *Metric) SetCourseID(courseID string) {
	if courseID != "" {
		this.Attributes[COURSE_ID_KEY] = courseID
	}
}

func (this *Metric) SetEndpoint(endpoint string) {
	if endpoint != "" {
		this.Attributes[ENDPOINT_KEY] = endpoint
	}
}

func (this *Metric) SetLocator(locator string) {
	if locator != "" {
		this.Attributes[LOCATOR_KEY] = locator
	}
}

func (this *Metric) SetSender(sender string) {
	if sender != "" {
		this.Attributes[SENDER_KEY] = sender
	}
}

func (this *Metric) SetTaskType(taskType string) {
	if taskType != "" {
		this.Attributes[TASK_TYPE_KEY] = taskType
	}
}

func (this *Metric) SetUserEmail(email string) {
	if email != "" {
		this.Attributes[USER_EMAIL_KEY] = email
	}
}
