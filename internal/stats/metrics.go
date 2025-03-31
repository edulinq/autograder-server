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
	Unknown_Metric_Attribute_Key MetricAttribute = ""
	Assignment_ID_Key                            = "assignment"
	Analysis_Type_Key                            = "analysis-type"
	Course_ID_Key                                = "course"
	Endpoint_Key                                 = "endpoint"
	Locator_Key                                  = "locator"
	Sender_Key                                   = "sender"
	Task_Type_Key                                = "task-type"
	User_Email_Key                               = "user"
)

// Values for the type field inside of Metric and Query.
const (
	Unknown_Metric_Attribute_Type MetricType = ""
	API_Request_Stats_Type                   = "api-request"
	Code_Analysis_Time_Stats_Type            = "code-analysis-time"
	Grading_Time_Stats_Type                  = "grading-time"
	Task_Time_Stats_Type                     = "task-time"
)

const ATTRIBUTES_KEY = "attributes"

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

func (this *Metric) addAttr(attrs *[]*log.Attr, key MetricAttribute) {
	val, ok := this.Attributes[key]
	if ok {
		*attrs = append(*attrs, log.NewAttr(string(key), val))
	}
}

func (this *Metric) LogValue() []*log.Attr {
	attrs := []*log.Attr{
		log.NewAttr("metric-type", this.Type),
	}

	this.addAttr(&attrs, Assignment_ID_Key)
	this.addAttr(&attrs, Analysis_Type_Key)
	this.addAttr(&attrs, Course_ID_Key)
	this.addAttr(&attrs, Endpoint_Key)
	this.addAttr(&attrs, Locator_Key)
	this.addAttr(&attrs, Sender_Key)
	this.addAttr(&attrs, Task_Type_Key)
	this.addAttr(&attrs, User_Email_Key)

	return attrs
}

func (this *Metric) Validate() error {
	if this == nil {
		return fmt.Errorf("No metric was given.")
	}

	if this.Type == Unknown_Metric_Attribute_Type {
		return fmt.Errorf("Metric attribute was not set.")
	}

	if this.Timestamp.IsZero() {
		return fmt.Errorf("Metric timestamp was not set.")
	}

	if this.Attributes == nil {
		this.Attributes = make(map[MetricAttribute]any)
	}

	for key, value := range this.Attributes {
		if key == Unknown_Metric_Attribute_Key {
			return fmt.Errorf("Metric attribute key was empty.")
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

func (this *Metric) setAttrIfNotEmpty(key MetricAttribute, value string) {
	if value != "" {
		this.Attributes[key] = value
	}
}

func (this *Metric) SetAssignmentID(id string) {
	this.setAttrIfNotEmpty(Assignment_ID_Key, id)
}

func (this *Metric) SetAnalysisType(analysis string) {
	this.setAttrIfNotEmpty(Analysis_Type_Key, analysis)
}

func (this *Metric) SetCourseID(courseID string) {
	this.setAttrIfNotEmpty(Course_ID_Key, courseID)
}

func (this *Metric) SetEndpoint(endpoint string) {
	this.setAttrIfNotEmpty(Endpoint_Key, endpoint)
}

func (this *Metric) SetLocator(locator string) {
	this.setAttrIfNotEmpty(Locator_Key, locator)
}

func (this *Metric) SetSender(sender string) {
	this.setAttrIfNotEmpty(Sender_Key, sender)
}

func (this *Metric) SetTaskType(taskType string) {
	this.setAttrIfNotEmpty(Task_Type_Key, taskType)
}

func (this *Metric) SetUserEmail(email string) {
	this.setAttrIfNotEmpty(User_Email_Key, email)
}
