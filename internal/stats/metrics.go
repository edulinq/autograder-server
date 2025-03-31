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
	MetricAttributeUnknown      MetricAttribute = ""
	MetricAttributeAssignmentID                 = "assignment"
	MetricAttributeAnalysisType                 = "analysis-type"
	MetricAttributeCourseID                     = "course"
	MetricAttributeEndpoint                     = "endpoint"
	MetricAttributeLocator                      = "locator"
	MetricAttributeSender                       = "sender"
	MetricAttributeTaskType                     = "task-type"
	MetricAttributeUserEmail                    = "user"
)

// Values for the type field inside of Metric and Query.
const (
	MetricTypeUnknown          MetricType = ""
	MetricTypeAPIRequest                  = "api-request"
	MetricTypeCodeAnalysisTime            = "code-analysis-time"
	MetricTypeGradingTime                 = "grading-time"
	MetricTypeTaskTime                    = "task-time"
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

// Map for quick existence check (empty value does not count).
var knownMetricTypes = map[MetricType]bool{
	MetricTypeUnknown:          false,
	MetricTypeAPIRequest:       true,
	MetricTypeCodeAnalysisTime: true,
	MetricTypeGradingTime:      true,
	MetricTypeTaskTime:         true,
}

// Map for quick existence check (empty value does not count).
var knownMetricAttributes = map[MetricAttribute]bool{
	MetricAttributeUnknown:      false,
	MetricAttributeAssignmentID: true,
	MetricAttributeAnalysisType: true,
	MetricAttributeCourseID:     true,
	MetricAttributeEndpoint:     true,
	MetricAttributeLocator:      true,
	MetricAttributeSender:       true,
	MetricAttributeTaskType:     true,
	MetricAttributeUserEmail:    true,
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
		if key == MetricAttributeUnknown {
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
	if metricType == MetricTypeUnknown {
		return fmt.Errorf("Metric type was not set.")
	}

	if !knownMetricTypes[metricType] {
		return fmt.Errorf("Unknown metric type: '%v'.", metricType)
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

func (this *Metric) setAttributeIfNotEmpty(key MetricAttribute, value string) {
	if value != "" {
		this.Attributes[key] = value
	}
}

func (this *Metric) SetAssignmentID(id string) {
	this.setAttributeIfNotEmpty(MetricAttributeAssignmentID, id)
}

func (this *Metric) SetAnalysisType(analysis string) {
	this.setAttributeIfNotEmpty(MetricAttributeAnalysisType, analysis)
}

func (this *Metric) SetCourseID(courseID string) {
	this.setAttributeIfNotEmpty(MetricAttributeCourseID, courseID)
}

func (this *Metric) SetEndpoint(endpoint string) {
	this.setAttributeIfNotEmpty(MetricAttributeEndpoint, endpoint)
}

func (this *Metric) SetLocator(locator string) {
	this.setAttributeIfNotEmpty(MetricAttributeLocator, locator)
}

func (this *Metric) SetSender(sender string) {
	this.setAttributeIfNotEmpty(MetricAttributeSender, sender)
}

func (this *Metric) SetTaskType(taskType string) {
	this.setAttributeIfNotEmpty(MetricAttributeTaskType, taskType)
}

func (this *Metric) SetUserEmail(email string) {
	this.setAttributeIfNotEmpty(MetricAttributeUserEmail, email)
}
