package stats

import (
	"fmt"

	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/log"
	"github.com/edulinq/autograder/internal/timestamp"
)

type MetricAttribute string
type MetricType string

// Keys for the attributes field inside of Metric and Query.
const (
	UnknownMetricAttributeKey MetricAttribute = ""
	AssignmentIDKey                           = "assignment"
	AnalysisTypeKey                           = "analysis-type"
	CourseIDKey                               = "course"
	EndpointKey                               = "endpoint"
	LocatorKey                                = "locator"
	SenderKey                                 = "sender"
	TaskTypeKey                               = "task-type"
	UserEmailKey                              = "user"
)

// Values for the type field inside of Metric and Query.
const (
	UnknownMetricAttributeType MetricType = ""
	APIRequestStatsType                   = "api-request"
	CodeAnalysisTimeStatsType             = "code-analysis-time"
	GradingTimeStatsType                  = "grading-time"
	TaskTimeStatsType                     = "task-time"
)

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

var knownMetricTypes = map[MetricType]bool{
	UnknownMetricAttributeType: true,
	APIRequestStatsType:        true,
	CodeAnalysisTimeStatsType:  true,
	GradingTimeStatsType:       true,
	TaskTimeStatsType:          true,
}

var knownMetricAttributes = map[MetricAttribute]bool{
	UnknownMetricAttributeKey: true,
	AssignmentIDKey:           true,
	AnalysisTypeKey:           true,
	CourseIDKey:               true,
	EndpointKey:               true,
	LocatorKey:                true,
	SenderKey:                 true,
	TaskTypeKey:               true,
	UserEmailKey:              true,
}

func (this Metric) GetTimestamp() timestamp.Timestamp {
	return this.Timestamp
}

func (this *Metric) LogValue() []*log.Attr {
	attrs := []*log.Attr{
		log.NewAttr("metric-type", this.Type),
	}

	for key, value := range this.Attributes {
		attrs = append(attrs, log.NewAttr(string(key), value))
	}

	return attrs
}

func (this *Metric) Validate() error {
	if this == nil {
		return fmt.Errorf("No metric was given.")
	}

	err := validateType(this.Type)
	if err != nil {
		return err
	}

	if this.Timestamp.IsZero() {
		return fmt.Errorf("Metric timestamp was not set.")
	}

	if this.Attributes == nil {
		this.Attributes = make(map[MetricAttribute]any)
	}

	return validateAttributeMap(this.Attributes)
}

func validateAttributeMap(attributes map[MetricAttribute]any) error {
	if attributes == nil {
		return nil
	}

	for key, value := range attributes {
		if key == UnknownMetricAttributeKey {
			return fmt.Errorf("Attribute key was empty.")
		}

		if !knownMetricAttributes[key] {
			return fmt.Errorf("Invalid attribute: '%v'.", key)
		}

		if value == nil || value == "" {
			return fmt.Errorf("Attribute value was empty for key '%v'.", key)
		}
	}

	return nil
}

func validateType(metricType MetricType) error {
	if metricType == UnknownMetricAttributeType {
		return fmt.Errorf("Type was not set.")
	}

	if !knownMetricTypes[metricType] {
		return fmt.Errorf("Invalid metric type: '%v'.", metricType)
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

func (this *Metric) setAttrIfNotEmpty(key MetricAttribute, value string) {
	if value != "" {
		this.Attributes[key] = value
	}
}

func (this *Metric) SetAssignmentID(id string) {
	this.setAttrIfNotEmpty(AssignmentIDKey, id)
}

func (this *Metric) SetAnalysisType(analysis string) {
	this.setAttrIfNotEmpty(AnalysisTypeKey, analysis)
}

func (this *Metric) SetCourseID(courseID string) {
	this.setAttrIfNotEmpty(CourseIDKey, courseID)
}

func (this *Metric) SetEndpoint(endpoint string) {
	this.setAttrIfNotEmpty(EndpointKey, endpoint)
}

func (this *Metric) SetLocator(locator string) {
	this.setAttrIfNotEmpty(LocatorKey, locator)
}

func (this *Metric) SetSender(sender string) {
	this.setAttrIfNotEmpty(SenderKey, sender)
}

func (this *Metric) SetTaskType(taskType string) {
	this.setAttrIfNotEmpty(TaskTypeKey, taskType)
}

func (this *Metric) SetUserEmail(email string) {
	this.setAttrIfNotEmpty(UserEmailKey, email)
}
