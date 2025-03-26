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
	ATTRIBUTE_KEY_TASK     = "task-type"
	ATTRIBUTE_KEY_ANALYSIS = "analysis-type"
	COURSE_ID              = "course"
	TYPE                   = "type"
	ASSIGNMENT_ID          = "assignment"
	USER_EMAIL             = "user"
	VALUE                  = "value"
	TASK_TYPE              = "task-type"
)

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

// Store a course metric without blocking (unless this is running in test mode, then it will block).
// Course ID is required, and all provided IDs should already be validated.
func AsyncStoreCourseMetric(metric *BaseMetric) {
	if metric == nil {
		return
	}

	if metric.Attributes == nil || metric.Attributes[COURSE_ID] == nil || metric.Attributes[COURSE_ID] == "" {
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
	attributes := map[string]any{
		TYPE:  CourseMetricTypeGradingTime,
		VALUE: uint64((endTime - startTime).ToMSecs()),
	}

	if userEmail != "" {
		attributes[USER_EMAIL] = userEmail
	}

	if courseID != "" {
		attributes[COURSE_ID] = courseID
	}

	if assignmentID != "" {
		attributes[ASSIGNMENT_ID] = assignmentID
	}

	metric := &BaseMetric{
		Timestamp:  startTime,
		Attributes: attributes,
	}

	AsyncStoreCourseMetric(metric)
}

func AsyncStoreCourseTaskTime(startTime timestamp.Timestamp, endTime timestamp.Timestamp, courseID string, assignmentID string, userEmail string, taskType string) {
	attributes := map[string]any{
		TYPE:      CourseMetricTypeTaskTime,
		TASK_TYPE: taskType,
		VALUE:     uint64((endTime - startTime).ToMSecs()),
	}

	if userEmail != "" {
		attributes[USER_EMAIL] = userEmail
	}

	if courseID != "" {
		attributes[COURSE_ID] = courseID
	}

	if assignmentID != "" {
		attributes[ASSIGNMENT_ID] = assignmentID
	}

	metric := &BaseMetric{
		Timestamp:  startTime,
		Attributes: attributes,
	}

	AsyncStoreCourseMetric(metric)
}
